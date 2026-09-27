SHELL := /bin/bash
GOPATH ?= $(shell go env GOPATH)
export PATH := /opt/homebrew/bin:$(GOPATH)/bin:$(PATH)

PROTOC ?= $(shell which protoc 2>/dev/null || echo /opt/homebrew/bin/protoc)
SWAG ?= $(shell which swag 2>/dev/null || echo $(GOPATH)/bin/swag)
GOLANGCI_LINT ?= $(shell which golangci-lint 2>/dev/null || echo $(GOPATH)/bin/golangci-lint)

.PHONY: all proto swag lint test vet build run clean

all: proto swag lint vet test build

## proto: Compile protobuf definitions into Go stubs
proto:
	@echo "==> Compiling protobuf definitions..."
	@mkdir -p proto/gen/go/v1
	$(PROTOC) --go_out=. --go_opt=module=github.com/moolroop \
	          --go-grpc_out=. --go-grpc_opt=module=github.com/moolroop \
	          api/proto/v1/identity.proto
	@mkdir -p gen
	@if [ ! -L gen/v1 ] && [ ! -d gen/v1 ]; then \
		ln -s ../proto/gen/go/v1 gen/v1; \
	fi

## swag: Auto-generate Swagger/OpenAPI documentation
swag:
	@echo "==> Generating Swagger documentation..."
	$(SWAG) init -g cmd/server/main.go -o docs/ --parseDependency --parseInternal

## lint: Run golangci-lint inspection
lint:
	@echo "==> Running golangci-lint..."
	$(GOLANGCI_LINT) run ./...

## test: Run unit test suites with race detector
test:
	@echo "==> Running tests with race detector..."
	go test -v -race ./...

## vet: Run go vet inspection
vet:
	@echo "==> Running go vet..."
	go vet ./...

## build: Build the dual-server binary
build:
	@echo "==> Building server binary..."
	@mkdir -p bin
	go build -o bin/server ./cmd/server

## run: Execute the service
run: build
	@echo "==> Running MoolRoop service..."
	./bin/server

## clean: Clean generated binaries
clean:
	@echo "==> Cleaning build artifacts..."
	rm -rf bin/
