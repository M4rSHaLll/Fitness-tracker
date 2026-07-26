DATABASE_URL ?= postgres://fitness:fitness@localhost:55432/fitness_tracker?sslmode=disable
APP_BINARY ?= bin/fitness-tracker

.PHONY: run run-postgres test build db-up db-down migrate-up migrate-down

run:
	go run ./cmd/telegram

run-postgres:
	STORAGE_DRIVER=postgres DATABASE_URL="$(DATABASE_URL)" go run ./cmd/telegram

test:
	go test ./...

build:
	go build -o $(APP_BINARY) ./cmd/telegram

db-up:
	docker compose up -d postgres

db-down:
	docker compose down

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down
