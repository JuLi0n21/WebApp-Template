package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/juli0n21/service/internal/db"
	api "github.com/juli0n21/service/proto"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) Login(ctx context.Context, req *api.LoginRequest) (*api.LoginResponse, error) {
	user, err := s.queries.GetUserByUsernameOrEmail(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if (user == db.User{}) {
		return nil, errors.New("invalid username/email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username/email or password")
	}

	tokenString, err := createJWTToken(user.Uuid.String(), user.Username)
	if err != nil {
		return nil, err
	}

	return &api.LoginResponse{
		Token: tokenString,
	}, nil
}

func (s *Server) Register(ctx context.Context, username, email, password string) (string, error) {
	existingUserByUsername, err := s.queries.GetUserByUsernameOrEmail(ctx, username)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if existingUserByUsername != (db.User{}) {
		return "", errors.New("username already taken")
	}

	existingUserByEmail, err := s.queries.GetUserByUsernameOrEmail(ctx, email)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if existingUserByEmail != (db.User{}) {
		return "", errors.New("email already registered")
	}

	hashedPasswordBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	hashedPassword := string(hashedPasswordBytes)

	newUUID, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate user uuid: %w", err)
	}

	err = s.queries.UpsertUser(ctx, db.UpsertUserParams{
		Uuid:           newUUID,
		Username:       username,
		Email:          email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return createJWTToken(newUUID.String(), username)
}

func createJWTToken(userID, username string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"name": username,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

type User struct {
	UUID           uuid.UUID `db:"uuid"`
	Username       string    `db:"username"`
	Email          string    `db:"email"`
	HashedPassword string    `db:"hashed_password"`
	Salt           string    `db:"salt"`
}
