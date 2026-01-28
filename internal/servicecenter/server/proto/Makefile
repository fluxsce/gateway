# Makefile for generating gRPC code from proto files (Linux/macOS)
# For Windows, please use proto.bat

# Protocol Buffers compiler
PROTOC := protoc

# Go plugins
PROTOC_GEN_GO := $(shell go env GOPATH)/bin/protoc-gen-go
PROTOC_GEN_GO_GRPC := $(shell go env GOPATH)/bin/protoc-gen-go-grpc

# Proto files
PROTO_FILES := $(wildcard *.proto)

# Output directory
OUT_DIR := .

.PHONY: all install clean

all: install generate

# Install protoc plugins
install:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate gRPC code
generate:
	@echo "Generating gRPC code from proto files..."
	$(PROTOC) --go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILES)
	@echo "Code generation completed!"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f *.pb.go *_grpc.pb.go
	@echo "Clean completed!"

