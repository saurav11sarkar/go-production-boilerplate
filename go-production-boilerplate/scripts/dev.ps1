$ErrorActionPreference = "Stop"
if (-not (Test-Path ".env")) {
  Copy-Item ".env.example" ".env"
  Write-Host "Created .env from .env.example"
}
docker compose up -d postgres
go run ./cmd/migrate up
go run ./cmd/seed
go run ./cmd/api
