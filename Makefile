# HAWKLING Makefile for development

# Variables
BINARY_NAME = hawkling
BUILD_DIR = build
CMD_DIR = cmd

# Go commands
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOVET = $(GOCMD) vet
GOMOD = $(GOCMD) mod
GOFMT = $(GOCMD) fmt
GOLINT = golangci-lint
GOTIDY = $(GOMOD) tidy

# Targets
.PHONY: all build clean test vet fmt lint tidy run-hawkling

all: tidy fmt vet lint test build

build: build-release

build-dev:
	go build -o build/hawkling-dev ./cmd/hawkling

# リリースビルド（サイズ最適化）
build-release:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o build/hawkling ./cmd/hawkling

clean:
	rm -rf $(BUILD_DIR)

test:
	$(GOTEST) -v ./...

vet:
	$(GOVET) ./...

fmt:
	$(GOFMT) ./...

lint:
	$(GOLINT) run

tidy:
	$(GOTIDY)

run: build-hawkling
	./$(BUILD_DIR)/$(BINARY_NAME_HAWKLING)
