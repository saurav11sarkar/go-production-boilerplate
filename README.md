# Go Production Boilerplate

`net/http` + Go standard `database/sql` + PostgreSQL (`pgx` stdlib driver) দিয়ে একটি **simple, learning-friendly, production-oriented modular monolith**।

এটা এমনভাবে বানানো হয়েছে যাতে আপনি code follow করে Go backend architecture শিখতে পারেন এবং নতুন project শুরু করার base হিসেবেও reuse করতে পারেন।

## Included

- Go `net/http` routing
- PostgreSQL through standard `database/sql` (`*sql.DB`) + pgx stdlib driver
- migrations + admin seed
- handler -> service -> repository architecture
- DTO + domain model
- access JWT
- refresh token with HttpOnly cookie
- refresh token rotation + refresh/user database row locks
- token-version based immediate session invalidation
- logout + logout all
- user/admin RBAC
- email verification
- forgot/reset password
- bcrypt password hashing
- consistent errors/responses
- request ID + structured `slog` logging
- panic recovery
- CORS + security headers
- request timeout context
- in-memory IP rate limit
- pagination + search + filter + sort whitelist
- local image upload for development
- Cloudinary provider for production
- avatar update/delete
- log mailer for development
- SMTP mailer for production
- graceful shutdown
- health/readiness endpoints
- Docker + docker-compose API/PostgreSQL
- unit tests + GitHub Actions CI

## Requirement

Recommended: **Go 1.27.1** and Docker Desktop. Go 1.27.1 is the current patch release used by the CI file.


## Database choice: `database/sql` + pgx driver

এই boilerplate application code-এ `pgxpool.Pool` ব্যবহার করে না। Standard library-এর `*sql.DB` ব্যবহার করে:

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)

db, err := sql.Open("pgx", databaseURL)
if err != nil {
    return err
}

db.SetMaxOpenConns(20)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(time.Hour)
db.SetConnMaxIdleTime(30 * time.Minute)

if err := db.PingContext(ctx); err != nil {
    return err
}
```

এখানে `database/sql` abstraction/pool API, আর `pgx` শুধু PostgreSQL driver। ফলে repository-তে আপনি পরিচিত standard API ব্যবহার করবেন:

```text
QueryRowContext
QueryContext
ExecContext
BeginTx
sql.Tx
sql.ErrNoRows
```

পরে MySQL/SQLite শেখার সময় একই `database/sql` concepts কাজে লাগবে; তবে SQL syntax এবং driver database অনুযায়ী বদলাতে পারে। `*sql.DB` নিজেই concurrency-safe connection pool, তাই এর উপরে আলাদা custom pool বানানোর দরকার নেই।

Development database pool config:

```env
DB_MAX_OPEN_CONNS=20
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=30m
```

## 1. Extract and enter project

```powershell
cd go-production-boilerplate
```

## 2. Create `.env`

PowerShell:

```powershell
Copy-Item .env.example .env
```

Then keep the development values initially. Change secrets before production.

## 3. Download Go modules

```powershell
go mod tidy
```

This also creates/updates `go.sum`. Commit `go.sum` to Git after the first successful tidy.

## 4. Start PostgreSQL

```powershell
docker compose up -d postgres
```

Default database:

```text
host: localhost
port: 5432
database: go_production
user: postgres
password: postgres
```

## 5. Run migrations

```powershell
go run ./cmd/migrate up
```

## 6. Seed admin

`.env` default development admin:

```text
admin@example.com
ChangeMe123!
```

Run:

```powershell
go run ./cmd/seed
```

Change this password immediately for any non-local environment.

## 7. Start API

```powershell
go run ./cmd/api
```

Then:

```text
API:     http://localhost:8080
health:  http://localhost:8080/healthz
ready:   http://localhost:8080/readyz
```

## One-command-ish Windows development

After Docker Desktop is running:

```powershell
./scripts/dev.ps1
```

It creates `.env` if missing, starts PostgreSQL, migrates, seeds admin, then starts API.

## Full Docker flow

If you want API + PostgreSQL both inside Docker:

```powershell
Copy-Item .env.example .env
docker compose up -d postgres
docker compose run --rm migrate
docker compose run --rm seed
docker compose up -d api
```

Then check:

```text
http://localhost:8080/healthz
http://localhost:8080/readyz
```

## Run tests

```powershell
go test ./...
```

Race detector:

```powershell
go test -race ./...
```

Static checks:

```powershell
go vet ./...
go fmt ./...
```

After the API is running and the default admin is seeded, run the integration smoke flow:

```powershell
./tests/integration/admin-flow.ps1
```

It tests health, readiness, admin login, authenticated profile, admin user listing, and refresh-token rotation.

## First login test

Use admin because seed marks it verified.

```http
POST http://localhost:8080/api/v1/auth/login
Content-Type: application/json
```

```json
{
  "email": "admin@example.com",
  "password": "ChangeMe123!"
}
```

Response contains the short-lived access JWT. Refresh token is set as an **HttpOnly cookie**.

Copy `accessToken`, then call:

```http
GET http://localhost:8080/api/v1/admin/users
Authorization: Bearer YOUR_ACCESS_TOKEN
```

## Register test

```json
{
  "name": "Saurav Sarkar",
  "email": "saurav@example.com",
  "password": "StrongPass123!"
}
```

With `MAIL_PROVIDER=log`, the verification URL appears in API terminal logs. Copy the `token` query value and call:

```http
POST /api/v1/auth/verify-email
Content-Type: application/json
```

```json
{
  "token": "PASTE_TOKEN_FROM_LOG"
}
```

A real frontend can read the token from the `/verify-email?token=...` link and make this POST request. This avoids consuming verification tokens from automatic email-link scanners.

## Refresh test

After login, Postman/Bruno/Apidog should keep the `refresh_token` cookie automatically.

Call:

```http
POST /api/v1/auth/refresh
```

The old refresh token is revoked and a new one replaces it. Refresh tokens are bound to `users.token_version`, so password change/reset and logout-all invalidate previous access **and** refresh sessions immediately.

Browser apps should prefer the HttpOnly cookie flow.

For a mobile/non-browser client that needs the raw refresh token in JSON, set:

```env
REFRESH_TOKEN_IN_BODY=true
```

Then login and refresh responses also include `refreshToken`, and `/auth/refresh`/`/auth/logout` can accept:

```json
{
  "refreshToken": "RAW_REFRESH_TOKEN"
}
```

Keep the default `false` for browser applications.

## Avatar upload

Development default:

```env
STORAGE_PROVIDER=local
```

Upload:

```http
POST /api/v1/users/me/avatar
Authorization: Bearer ACCESS_TOKEN
Content-Type: multipart/form-data
```

Field name:

```text
image
```

Allowed: JPEG, PNG, WebP. Default max: 5 MB.

Files are available under `/uploads/...` locally.

### Switch to Cloudinary

```env
STORAGE_PROVIDER=cloudinary
CLOUDINARY_CLOUD_NAME=...
CLOUDINARY_API_KEY=...
CLOUDINARY_API_SECRET=...
```

No service/handler code change is needed because both providers implement the same storage interface.

## Email production setup

Local:

```env
MAIL_PROVIDER=log
```

Production SMTP:

```env
MAIL_PROVIDER=smtp
MAIL_FROM_NAME=My App
MAIL_FROM=noreply@example.com
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=...
SMTP_PASSWORD=...
```

## Production auth values

Minimum:

```env
APP_ENV=production
JWT_ACCESS_SECRET=USE_A_LONG_RANDOM_SECRET
REFRESH_COOKIE_SECURE=true
PUBLIC_BASE_URL=https://api.example.com
FRONTEND_URL=https://example.com
CORS_ALLOWED_ORIGINS=https://example.com,https://admin.example.com
```

Always run behind HTTPS in production. `APP_ENV=production` intentionally refuses insecure refresh cookies, the development log-mailer, placeholder JWT secrets, and non-HTTPS public/frontend URLs. Use `.env.production.example` as the production checklist starting point.

`REFRESH_COOKIE_SAMESITE=lax` is the safe default for normal same-site deployments. If your frontend and API are truly cross-site and you set it to `none`, `REFRESH_COOKIE_SECURE=true` is mandatory.

Set `TRUST_PROXY=true` **only** behind a trusted reverse proxy/load balancer that overwrites forwarding headers; otherwise clients could spoof forwarded IP/protocol headers.

## Folder map

```text
cmd/
  api/          HTTP executable
  migrate/      migration executable
  seed/         admin seed executable

internal/
  app/          dependency wiring, routes, server
  config/       env configuration
  database/     database/sql connection pool + transaction helper
  domain/       core entities
  dto/          request DTOs
  repository/   SQL/PostgreSQL access
  service/      business logic
  handler/      HTTP handlers
  auth/         JWT + auth context
  middleware/   request pipeline
  httpx/        JSON response/decode/error
  apperror/     app error types/codes
  query/        pagination/filter/search/sort
  storage/      local/Cloudinary
  mailer/       log/SMTP
  security/     bcrypt + secure random tokens

migrations/     SQL up/down migrations
seeds/          seed documentation
scripts/        local helper scripts
docs/           architecture/auth/API explanation
templates/      example email templates
```

## The rule to remember

```text
Handler    = HTTP
Service    = business rule
Repository = database
Middleware = cross-cutting request logic
Domain     = core data/types
DTO        = external request shape
```

A handler should not contain SQL. A repository should not know about `http.ResponseWriter`. A service should not decide HTTP status codes directly except through application errors returned to the HTTP layer.

## Pagination/search/filter/sort example

```text
GET /api/v1/admin/users
  ?page=1
  &limit=20
  &search=saurav
  &role=user
  &isVerified=true
  &isActive=true
  &sortBy=createdAt
  &sortOrder=desc
```

`limit` is capped at 100. Sort fields are mapped through a whitelist before SQL is built.

## Important production note: rate limiting

The included rate limiter is process-memory based. It is suitable for learning, a single API process, and a normal single-VPS deployment. Sensitive auth endpoints (`register`, `login`, verification resend, forgot/reset password) also have a separate stricter IP limit.

If you horizontally scale to multiple API instances, use Redis (or gateway/load-balancer rate limiting) so every instance shares the same counters.

## Docs

Read in this order:

1. `docs/ARCHITECTURE.md`
2. `docs/DATABASE_SQL.md`
3. `docs/AUTH_FLOW.md`
4. `docs/API.md`
5. `internal/database/postgres.go`
6. `internal/repository/user_repository.go`
7. `internal/app/routes.go`
8. `internal/handler/auth.go`
9. `internal/service/auth_service.go`

এই order-এ পড়লে request কোথা দিয়ে যায় সেটা সবচেয়ে সহজে বুঝবেন।

