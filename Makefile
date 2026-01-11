.PHONY: dev docker-up docker-down test test-backend test-frontend generate migrate-up migrate-down

# Development
dev:
	docker-compose up -d db mailhog
	@echo "Starting backend (Air hot reload)..."
	cd backend && air &
	@echo "Starting frontend (Vite dev server)..."
	cd frontend && npm run dev

# Docker commands
docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

# Database
migrate-up:
	cd backend && migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	cd backend && migrate -path db/migrations -database "$(DATABASE_URL)" down 1

# Code generation
generate:
	cd backend && sqlc generate

# Testing
test: test-backend test-frontend

test-backend:
	cd backend && go test ./...

test-frontend:
	cd frontend && npm test

# Build
build-backend:
	cd backend && go build -o ./bin/server ./cmd/server

build-frontend:
	cd frontend && npm run build

# Clean
clean:
	rm -rf backend/tmp
	rm -rf backend/bin
	rm -rf frontend/dist
