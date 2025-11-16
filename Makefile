.PHONY: build up down migrate clean run

build:
	go build -o bin/pr-reviewer ./cmd/server

up: build
	docker-compose up -d --build

down:
	docker-compose down

clean:
	docker-compose down -v
	rm -f bin/pr-reviewer

migrate:
	docker-compose exec db psql -U postgres -d prdb -f /migrations/0001_init.sql

run: up migrate
	docker-compose logs -f app
