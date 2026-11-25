.PHONY: run build swag up down logs k6

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api

swag:
	swag init -g cmd/api/main.go -o docs --parseInternal

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f api

k6:
	k6 run -e BASE_URL=$${BASE_URL:-http://localhost:8080} k6/load-test.js
