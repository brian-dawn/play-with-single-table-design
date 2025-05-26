.PHONY: up down build test run clean all

# Default target
all: build test

# Start Docker services
up:
	docker-compose up -d

# Stop Docker services
down:
	docker-compose down

# Build the application
build:
	go build -v .

watch:
	air

# Run tests
test: up
	go test -v ./...
	
# Run tests with coverage
test-coverage: up
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run the application
run: up build
	./LearnSingleTableDesign

# Clean build artifacts
clean:
	go clean
	rm -f coverage.out coverage.html
	rm -f LearnSingleTableDesign

# Show help
help:
	@echo "Available targets:"
	@echo "  up            - Start Docker services (DynamoDB Local and Admin)"
	@echo "  down          - Stop Docker services"
	@echo "  build         - Build the application"
	@echo "  watch         - Watch for changes and rerun the application, runs a proxy server on :8081"
	@echo "  test          - Run tests (starts Docker services first)"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  run           - Run the application (starts Docker services first)"
	@echo "  clean         - Clean build artifacts"
	@echo "  all           - Build and test (default)"
	@echo "  help          - Show this help message" 