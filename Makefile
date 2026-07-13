##for local testing, remove from prod

DSN      ?= host=localhost user=postgres password=password dbname=grocerics port=5432 sslmode=disable
MIGR_DIR := ./internal/migrate/migrations
GOOSE    := go run github.com/pressly/goose/v3/cmd/goose@v3.27.2 -dir $(MIGR_DIR) postgres "$(DSN)"
SWAG     := go run github.com/swaggo/swag/cmd/swag@v1.16.6

.PHONY: up down docs migrate migrate-down migrate-status test build

up: ## build + run the full stack (Postgres + API+ migratations)
	docker compose up --build

down: ## stop the stack
	docker compose down

docs: ## regenerate Swagger docs from annotations into docs/
	$(SWAG) init -g cmd/main.go -o docs

migrate: ## apply all migrations against $(DSN)
	$(GOOSE) up

migrate-down: ## roll back the last migration
	$(GOOSE) down

migrate-status: ## show migration status
	$(GOOSE) status

test: ## run all tests
	go test ./...

build: ## compile all packages
	go build ./...
