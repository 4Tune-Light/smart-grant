.PHONY: build run-* dev lint test proto migrate clean

MODULE := github.com/rizky/smart-grant
BUILD_DIR := bin

# Build all services
build:
	@echo "Building all services..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/api-gateway ./cmd/api-gateway
	go build -o $(BUILD_DIR)/backend ./cmd/backend
	go build -o $(BUILD_DIR)/worker ./cmd/worker
	@echo "Build complete: $(BUILD_DIR)/"

run-api-gateway:
	@echo "Starting API Gateway..."
	APP_NAME=smart-grant APP_ENV=development go run ./cmd/api-gateway

run-backend:
	@echo "Starting Backend Service..."
	APP_NAME=smart-grant APP_ENV=development go run ./cmd/backend

run-worker:
	@echo "Starting Worker..."
	APP_NAME=smart-grant APP_ENV=development go run ./cmd/worker

# Run all services via docker-compose
dev:
	@echo "Starting all services..."
	docker compose -f deploy/docker-compose.yml up --build

dev-down:
	docker compose -f deploy/docker-compose.yml down

dev-logs:
	docker compose -f deploy/docker-compose.yml logs -f

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v -race -count=1

test-short:
	@echo "Running short tests..."
	go test ./... -short -race -count=1

# Lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Generate protobuf stubs
proto:
	@echo "Generating protobuf stubs..."
	protoc --go_out=. --go-grpc_out=. proto/*.proto

# Migration helpers
migrate-up:
	@echo "Running migrations up..."
	go run ./scripts/migrate.go up

migrate-down:
	@echo "Running migrations down..."
	go run ./scripts/migrate.go down

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	@echo "Done."

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
