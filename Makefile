.PHONY: build test clean run docker-build docker-run

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Binary info
BINARY_NAME=audit-logs-monitoring
BINARY_PATH=./bin/$(BINARY_NAME)
MAIN_PATH=./cmd

# Build the application
build:
	$(GOBUILD) -o $(BINARY_PATH) $(MAIN_PATH)

# Test all packages
test:
	$(GOTEST) -v ./...

# Test with coverage
test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_PATH)
	rm -f coverage.out coverage.html

# Run the application
run: build
	$(BINARY_PATH)

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Lint the code
lint:
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):latest .

# Run Docker container
docker-run:
	docker run -p 8080:8080 $(BINARY_NAME):latest

# Development setup
dev-setup:
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint

# Format code
fmt:
	$(GOCMD) fmt ./...

# Vet code
vet:
	$(GOCMD) vet ./...

# All checks
check: fmt vet lint test