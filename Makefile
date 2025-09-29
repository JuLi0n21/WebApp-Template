GOOGLEAPIS_DIR = backend/googleapis
GATEWAY_DIR = backend/grpc-gateway/
GO_OUT_DIR = backend
SWAGGER_OUT_DIR = backend/gen/swagger
FRONEND_SWAGGER_OUT_DIR = frontend/api

PROTO_DIR = proto

PROTOC ?= protoc
PROTOC_GEN_GO ?= protoc-gen-go
PROTOC_GEN_GO_GRPC ?= protoc-gen-go-grpc
PROTOC_GEN_GRPC_GATEWAY ?= protoc-gen-grpc-gateway
PROTOC_GEN_OPENAPIV2 ?= protoc-gen-openapiv2
OPENAPI_GENERATOR ?= openapi-generator-cli

API_CLIENT_OUT_DIR = frontend/src/api-client

all: generate

generate:
	$(PROTOC) \
		--proto_path=. \
		--proto_path=$(GOOGLEAPIS_DIR) \
		--proto_path=$(GATEWAY_DIR) \
		--go_out=$(GO_OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(GO_OUT_DIR) --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=$(SWAGGER_OUT_DIR) \
		--openapiv2_opt allow_merge=true \
		--openapiv2_opt merge_file_name=api \
		$(PROTO_DIR)/*.proto

		cp $(SWAGGER_OUT_DIR)/api.swagger.json $(FRONEND_SWAGGER_OUT_DIR)/

api: $(SWAGGER_FILE)
	$(OPENAPI_GENERATOR) generate \
		-i $(SWAGGER_OUT_DIR)/api.swagger.json \
		-g typescript-axios \
		-o $(API_CLIENT_OUT_DIR) \
		--additional-properties=supportsES6=true,withSeparateModelsAndApi=true,modelPropertyNaming=original,apiPackage=apis,modelPackage=models

.PHONY: all generate clean

