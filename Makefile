.PHONY: build run test clean docker-up docker-down install-deps help

# Default target
help:
	@echo "Available targets:"
	@echo "  make build          - Build the API server"
	@echo "  make run            - Run the API server locally"
	@echo "  make test           - Run unit tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-up      - Start Docker Compose services"
	@echo "  make docker-down    - Stop Docker Compose services"
	@echo "  make docker-logs    - View Docker Compose logs"
	@echo "  make install-deps   - Download and install dependencies"
	@echo "  make fmt            - Format code with gofmt"
	@echo "  make lint           - Run golangci-lint (requires installation)"

# Build the application
build:
	go build -o api-server ./cmd/api
	@echo "✓ Build complete: api-server"

# Run the application locally
run:
	go run ./cmd/api

# Run unit tests
test:
	go test -v ./...

# Run tests with coverage report
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	@echo "✓ Coverage report generated: coverage.html"

# Run tests in watch mode (requires entr: brew install entr)
test-watch:
	find . -name "*.go" | entr -r make test

# Clean build artifacts
clean:
	rm -f api-server coverage.out coverage.html
	go clean ./...
	@echo "✓ Clean complete"

# Download and tidy dependencies
install-deps:
	go mod download
	go mod tidy
	@echo "✓ Dependencies installed"

# Format code
fmt:
	go fmt ./...
	@echo "✓ Code formatted"

# Run linter (requires: brew install golangci-lint)
lint:
	golangci-lint run ./...

# Docker: Build image
docker-build:
	docker build -f docker/Dockerfile -t api-server .
	@echo "✓ Docker image built: api-server"

# Docker: Start services with docker-compose
docker-up:
	docker-compose -f docker/docker-compose.yml up -d
	@echo "✓ Services started"
	@echo "  API:      http://localhost:8080"
	@echo "  Database: localhost:5432"
	@echo "  Cache:    localhost:6379"

# Docker: Stop services
docker-down:
	docker-compose -f docker/docker-compose.yml down
	@echo "✓ Services stopped"

# Docker: Remove all data
docker-clean:
	docker-compose -f docker/docker-compose.yml down -v
	@echo "✓ Services and volumes removed"

# Docker: View logs
docker-logs:
	docker-compose -f docker/docker-compose.yml logs -f

# Docker: View specific service logs
docker-logs-api:
	docker-compose -f docker/docker-compose.yml logs -f api

docker-logs-db:
	docker-compose -f docker/docker-compose.yml logs -f db

docker-logs-redis:
	docker-compose -f docker/docker-compose.yml logs -f redis

# Full development setup: install deps, build, and start docker services
dev-setup: install-deps build docker-up
	@echo "✓ Development environment ready"

# Run all checks: format, lint, test
check: fmt lint test
	@echo "✓ All checks passed"
