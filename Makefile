# Makefile for azkeyget development

.PHONY: build test lint clean install-tools fmt vet

# Build the binary
build:
	go build -ldflags="-s -w" -o azkeyget ./cmd/azkeyget

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/go-critic/go-critic/cmd/gocritic@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/mgechev/revive@latest

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Run go vet
vet:
	go vet ./...

# Run all linters
lint:
	golangci-lint run
	revive -config revive.toml ./...

# Run pre-commit hooks manually
pre-commit:
	pre-commit run --all-files

# Clean build artifacts
clean:
	rm -f azkeyget
	rm -f azkeyget_test
	go clean

# Run all checks (format, vet, lint, test)
check: fmt vet lint test

# Default target
all: fmt vet test build
