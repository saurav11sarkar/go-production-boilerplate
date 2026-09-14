#!/usr/bin/env sh
set -eu
[ -f .env ] || cp .env.example .env
docker compose up -d postgres
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
