.PHONY: build up rebuild down logs ps swagger deps

# Docker Compose helpers for the complete local stack.

build:
	docker compose build

up:
	docker compose up -d

rebuild:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

ps:
	docker compose ps

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g main.go -d cmd/api,internal/core/error,internal/inbound/http/adapter,internal/inbound/http/dto -o docs

deps:
	go mod download
