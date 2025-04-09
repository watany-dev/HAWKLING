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

build: build-hawkling

build-hawkling:
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME_HAWKLING) ./$(CMD_DIR)/hawkling

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
