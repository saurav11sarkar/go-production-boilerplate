.PHONY: run build test test-race fmt vet tidy migrate-up migrate-down seed docker-up docker-down

run:
	go run ./cmd/api

build:
	go build ./...

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
