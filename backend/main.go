package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"github.com/juli0n21/service/internal/db"
	api "github.com/juli0n21/service/proto"
	_ "github.com/lib/pq"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	Port        string
	GatewayPort string
	db          *sql.DB
	queries     *db.Queries
	Env         string

	api.UnimplementedServiceServer
}

var jwtSecret []byte

var Default_Limit = 100

func main() {
	_ = godotenv.Load()

	env := os.Getenv("ENV")
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://user:password@localhost:5432/database?sslmode=disable"
	}

	sqlDB, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("could not run migrate up: %v", err)
	}
	log.Println("Migrations ran successfully")

	queries := db.New(sqlDB)

	s := &Server{
		db:          sqlDB,
		queries:     queries,
		Port:        ":9090",
		GatewayPort: ":8080",
	}

	if env == "development" {
		_, err = s.Register(context.Background(), "username", "thisisanemail@web.de", "password")
		if err != nil {
			log.Println("Failed to register test user:", err)
		}
	}

	jwtSecretStr := os.Getenv("JWT_SECRET")
	if jwtSecretStr == "" {
		if env == "development" {
			jwtSecretStr = "default-dev-secret"
		} else {
			log.Fatal("JWT_SECRET environment variable is not set")
		}
	}
	jwtSecret = []byte(jwtSecretStr)

	log.Fatal(runGrpcAndGateway(s))
}

func runGrpcAndGateway(s *Server) error {
	grpcPort := s.Port
	gatewayPort := s.GatewayPort

	grpcLis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", grpcPort, err)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authUnaryInterceptor()))
	api.RegisterServiceServer(grpcServer, s)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwMux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = api.RegisterServiceHandlerFromEndpoint(ctx, gwMux, grpcPort, opts)
	if err != nil {
		return fmt.Errorf("failed to register grpc-gateway: %w", err)
	}

	mux := &http.ServeMux{}

	fileServer := http.FileServer(http.Dir("gen/swagger"))
	mux.Handle("/swagger/", http.StripPrefix("/swagger/", fileServer))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gwMux.ServeHTTP(w, r)
	}))

	httpServer := &http.Server{
		Addr:    gatewayPort,
		Handler: corsMiddleware(logRequests(authMiddleware(mux))),
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 2)

	go func() {
		log.Printf("Starting gRPC server on http://localhost%s", grpcPort)
		errChan <- grpcServer.Serve(grpcLis)
	}()

	go func() {
		log.Printf("Starting HTTP gateway server on http://localhost%s/swagger", gatewayPort)
		errChan <- httpServer.ListenAndServe()
	}()

	select {
	case <-stop:
		log.Println("Shutting down servers...")
		grpcServer.GracefulStop()
		httpServer.Shutdown(ctx)
		return nil
	case err := <-errChan:
		return err
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.HasPrefix(r.URL.Path, "/swagger") && os.Getenv("ENV") == "development" {
			next.ServeHTTP(w, r)
			return
		}

		if strings.HasPrefix(r.URL.Path, "/v1/auth/login") {
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/v1/health" {
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(auth, "Bearer ")
		_, err := validateToken(tokenString)
		if err != nil {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if info.FullMethod == "/horseshoe.HorseshoeService/Login" {
			return handler(ctx, req)
		}

		if info.FullMethod == "/horseshoe.HorseshoeService/HealthCheck" {
			return handler(ctx, req)
		}

		if strings.HasPrefix(info.FullMethod, "/horseshoe.HorseshoeService/Swagger") && os.Getenv("ENV") == "development" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		auth := md["authorization"]
		if len(auth) == 0 || !strings.HasPrefix(auth[0], "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "missing or invalid authorization header")
		}
		claims, err := validateToken(strings.TrimPrefix(auth[0], "Bearer "))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Optionally set claims into context for use in handlers
		newCtx := context.WithValue(ctx, "userClaims", claims)
		return handler(newCtx, req)
	}
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
	parser := jwt.NewParser(jwt.WithLeeway(5 * time.Second))

	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token or claims")
	}

	return claims, nil
}

func SafeLimit(limit int32) int32 {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	return limit
}
