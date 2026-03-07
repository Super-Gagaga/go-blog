.PHONY: help build run dev test lint clean migrate-up migrate-down docker-up docker-down

BINARY := bin/server
MAIN := ./cmd/server
DATABASE_URL ?= mysql://go_blog:go_blog@tcp(127.0.0.1:3306)/go_blog?multiStatements=true

help:
	@echo "build        - build the server binary"
	@echo "run          - run the built binary"
	@echo "dev          - run the server locally"
	@echo "test         - run tests"
	@echo "lint         - run golangci-lint"
	@echo "migrate-up   - apply database migrations"
	@echo "migrate-down - roll back one database migration"
	@echo "docker-up    - build and start the complete Docker stack"
	@echo "docker-down  - stop the Docker stack"

build:
	go build -o $(BINARY) $(MAIN)

run: build
	./$(BINARY)

dev:
	go run $(MAIN)

test:
	go test -v ./...

lint:
	golangci-lint run

clean:
	go clean
	rm -rf bin

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

docker-up:
	docker compose -f deploy/docker-compose.yml up -d --build

docker-down:
	docker compose -f deploy/docker-compose.yml down
