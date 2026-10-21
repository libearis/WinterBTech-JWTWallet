.PHONY: up down db-up db-down run-backend run-frontend test lint

up:
	docker compose up --build

down:
	docker compose down

db-up:
	docker compose -f docker-compose.db.yml up -d

db-down:
	docker compose -f docker-compose.db.yml down

run-backend:
	cd backend && go run ./cmd/api

run-frontend:
	cd frontend && npm run dev

test:
	cd backend && go test ./... -race -cover

lint:
	cd backend && go vet ./...
