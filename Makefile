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
# Terraform targets
.PHONY: tf-init tf-plan tf-apply tf-destroy tf-validate tf-fmt tf-output tf-state tf-cost

# Terraform: Initialize (dev environment)
tf-init:
	cd terraform/environments/dev && terraform init

# Terraform: Plan deployment (dev environment)
tf-plan:
	cd terraform/environments/dev && terraform plan -out=tfplan

# Terraform: Apply deployment (dev environment)
tf-apply:
	cd terraform/environments/dev && terraform apply tfplan

# Terraform: Destroy infrastructure (dev environment)
tf-destroy:
	cd terraform/environments/dev && terraform destroy

# Terraform: Validate configuration
tf-validate:
	terraform -chdir=terraform validate

# Terraform: Format code
tf-fmt:
	terraform fmt -recursive terraform/

# Terraform: View outputs
tf-output:
	cd terraform/environments/dev && terraform output

# Terraform: View state
tf-state:
	cd terraform/environments/dev && terraform state list

# Terraform: Estimate costs
tf-cost:
	@echo "Estimated costs:"
	@echo "Dev: $50-80/month"
	@echo "Prod: $200-400/month"

# Terraform: Plan prod deployment
tf-plan-prod:
	cd terraform/environments/prod && terraform init && terraform plan -out=tfplan

# Terraform: Apply prod deployment
tf-apply-prod:
	cd terraform/environments/prod && terraform apply tfplan

# ECR: Login
ecr-login:
	@aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $$(aws sts get-caller-identity --query Account --output text).dkr.ecr.us-east-1.amazonaws.com
	@echo "✓ Logged in to ECR"

# ECR: Build and push image
ecr-push: docker-build
	@export ECR_URL=$$(cd terraform/environments/dev && terraform output -raw ecr_repository_url 2>/dev/null || echo ""); \
	if [ -z "$$ECR_URL" ]; then \
		echo "❌ ECR repository not found. Run 'make tf-apply' first."; \
		exit 1; \
	fi; \
	docker tag api-server:latest $$ECR_URL:latest; \
	docker push $$ECR_URL:latest; \
	echo "✓ Image pushed to ECR"

# ECS: Update service (trigger redeployment)
ecs-deploy:
	aws ecs update-service \
		--cluster api-test-cluster \
		--service api-test-service \
		--force-new-deployment \
		--region us-east-1
	@echo "✓ ECS service updated"

# AWS: Get ALB URL
aws-alb-url:
	@cd terraform/environments/dev && terraform output alb_url 2>/dev/null || echo "ALB not deployed"

# AWS: View logs
aws-logs:
	aws logs tail /ecs/api-test-task --follow --region us-east-1

# AWS: Check service status
aws-status:
	@aws ecs describe-services \
		--cluster api-test-cluster \
		--services api-test-service \
		--region us-east-1 \
		--query 'services[0].{Status:status,DesiredCount:desiredCount,RunningCount:runningCount}' \
		--output table