.PHONY: run build test test-race fmt vet tidy migrate-up migrate-down seed docker-up docker-down dev clean

run:
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/seed ./cmd/seed

test:
	go test ./...

test-race:
	go test -race ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

tidy:
	go mod tidy

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

seed:
	go run ./cmd/seed

docker-up:
	docker compose up -d postgres

docker-down:
	docker compose down

dev:
	docker compose up -d postgres
	go run ./cmd/migrate up
	go run ./cmd/seed
	go run ./cmd/api

clean:
	go clean
