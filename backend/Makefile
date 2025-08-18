.PHONY: help dev build test lint clean migrate-up migrate-down swagger docker-build docker-run docker-stop

# Default target
help:
	@echo "Available commands:"
	@echo "  dev          - Run the application locally"
	@echo "  build        - Build the application"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  clean        - Clean build artifacts"
	@echo "  migrate-up   - Run database migrations"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  swagger      - Generate Swagger documentation"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"
	@echo "  docker-stop  - Stop Docker Compose services"

# Development
dev:
	@echo "Starting development server..."
	go run cmd/server/main.go

# Build
build:
	@echo "Building application..."
	go build -o bin/expense-tracker-backend cmd/server/main.go

# Test
test:
	@echo "Running tests..."
	go test -v ./...

# Lint
lint:
	@echo "Running linter..."
	golangci-lint run

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

# Database migrations
migrate-up:
	@echo "Running database migrations..."
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/expensedb?sslmode=disable" up

migrate-down:
	@echo "Rolling back database migrations..."
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/expensedb?sslmode=disable" down

# Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go -o docs

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker build -t expense-tracker-backend .

docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Setup development environment
setup: deps
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file..."; \
		cp .env.example .env; \
	fi
	@echo "Development environment setup complete!"

# Run all checks
check: lint test
	@echo "All checks passed!"

# Production build
prod-build:
	@echo "Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/expense-tracker-backend cmd/server/main.go 