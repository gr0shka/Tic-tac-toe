include .env
export

MIGRATION_DIR=./migration
DB_DSN=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DATABASE)?sslmode=disable

migrate-up:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_DSN)" up

migrate-down:
	goose -dir $(MIGRATION_DIR) postgres "$(DB_DSN)" down

test:
	go test -race ./...

test-cover:
	go test -coverprofile=covarage.out ./...

test-cover-html: test-cover
	go tool cover -html=covarage.out

clean:
	rm covarage.out

.PHONY: test test-cover test-cover-html