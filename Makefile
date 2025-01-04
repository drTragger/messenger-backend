# Load environment variables from .env file
include .env
export $(shell sed 's/=.*//' .env)

# Variables
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_DIR=db/migrations

# Targets
.PHONY: help run jwt-secret migrate-up migrate-down migrate-force create-migration check-env check-tools

## Show available commands
help:
	@echo "Available commands:"
	@echo "  run                - Run the application (includes migrations)"
	@echo "  jwt-secret         - Generate a new JWT secret"
	@echo "  migrate-up         - Apply all up migrations"
	@echo "  migrate-down       - Rollback the last migration"
	@echo "  migrate-force      - Force a specific migration version"
	@echo "  create-migration   - Create a new migration file"
	@echo "  check-env          - Verify the .env file is loaded"
	@echo "  check-tools        - Verify required tools are installed"

## Run the application (includes migrations)
run: check-env check-tools migrate-up
	go run cmd/main.go

## Generate a new JWT secret
jwt-secret: check-env
	./cmd/generate-secret

## Apply all up migrations
migrate-up: check-env check-tools
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) up

## Rollback the last migration
migrate-down: check-env check-tools
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) down 1

## Force a specific migration version
migrate-force: check-env check-tools
	@echo "Forcing migration to version: $(version)"
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) force $(version)

## Create a new migration file
create-migration: check-env
	@echo "Creating migration: $(name)"
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## Verify the .env file is loaded
check-env:
	@if [ ! -f .env ]; then \
		echo ".env file is missing. Please create it before running commands."; \
		exit 1; \
	fi

## Verify required tools are installed
check-tools:
	@command -v migrate >/dev/null 2>&1 || { echo "migrate is not installed. Please install it: https://github.com/golang-migrate/migrate"; exit 1; }