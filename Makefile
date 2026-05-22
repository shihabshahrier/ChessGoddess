.PHONY: dev test build lint migrate clean

# Start backend dev server (requires postgres + redis running)
dev:
	go run ./cmd/server

# Run all Go tests with race detector
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -1

# Build production binary
build:
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/chesslens ./cmd/server

# Run linter
lint:
	golangci-lint run ./...

# Run database migrations
migrate:
	@for f in migrations/*.sql; do echo "applying $$f"; psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -f "$$f"; done

# Start all services via Docker Compose
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# Frontend dev server
frontend:
	cd frontend && npm run dev

# Frontend build
frontend-build:
	cd frontend && npm run build

# Format Go code
fmt:
	gofmt -w .

clean:
	rm -rf bin/ coverage.out
