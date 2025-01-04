# Load environment variables from .env file
include .env
export $(shell sed 's/=.*//' .env)

# Variables
DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
MIGRATIONS_DIR=db/migrations

# Commands
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) down

migrate-force:
	migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) force $(version)

create-migration:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)