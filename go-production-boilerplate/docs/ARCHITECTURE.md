# Architecture Guide

এই project একটি **modular monolith**। Microservice না, আবার সব logic handler-এর ভেতরও না। শেখা এবং real production API—দুইটার জন্য balanced structure।

## Request flow

```text
Client
  -> global middleware
  -> auth / role middleware (protected routes only)
  -> handler
  -> service
  -> repository
  -> PostgreSQL
```

Response উল্টো পথে client-এ ফেরে।

## Layer responsibility

### `cmd/`
Executable entry points। এখানে business logic থাকবে না।

- `cmd/api`: HTTP API চালায়।
- `cmd/migrate`: database migration চালায়।
- `cmd/seed`: admin seed করে।

### `internal/app/`
Dependency wiring + routes + server lifecycle। Repository, service, handler এখানে connect হয়।

### `internal/domain/`
Core entity/types: `User`, `Role`, token models। HTTP বা database-specific behavior এখানে কম রাখা হয়।

### `internal/dto/`
Client request/response shapes। Database model সরাসরি request body হিসেবে ব্যবহার করা হয় না।

### `internal/handler/`
HTTP-specific কাজ:

- JSON/multipart read
- path/query parameter read
- service call
- status code/JSON response

Business rule handler-এ লিখবেন না।

### `internal/service/`
Business logic-এর main layer। যেমন login-এর সময় password compare, account active কিনা, email verified কিনা, token তৈরি ইত্যাদি।

Service ছোট interface consume করে, তাই unit test-এ fake/mock repository দেওয়া সহজ।

### `internal/database/`
Standard `database/sql` setup। `sql.Open("pgx", DATABASE_URL)` দিয়ে `*sql.DB` তৈরি হয়, pool tuning (`SetMaxOpenConns`, `SetMaxIdleConns`, connection lifetime) এবং reusable `sql.Tx` helper এখানে থাকে। `pgx` এখানে PostgreSQL driver; application/repository API standard `database/sql`।

### `internal/repository/`
PostgreSQL access। Repository `*sql.DB` ব্যবহার করে এবং `QueryRowContext`, `QueryContext`, `ExecContext`, `sql.ErrNoRows` দিয়ে SQL/row scanning handle করে। Transaction/row-locking flow-এ `*sql.Tx` ব্যবহার হয়।

### `internal/auth/`
JWT create/verify এবং authenticated user context।

### `internal/security/`
Password hash এবং cryptographically secure random token/hash helper।

### `internal/middleware/`
Common request pipeline:

```text
RequestID
-> Recover
-> Logger
-> SecurityHeaders
-> CORS
-> RateLimit
-> Timeout context
-> Auth (protected route)
-> Role (admin route)
```

### `internal/query/`
Pagination/filter/search/sort request parsing এবং meta response। Repository sort-field whitelist করে।

### `internal/storage/`
Storage abstraction। Development-এ local disk, production-এ Cloudinary। Service storage implementation সম্পর্কে জানে না।

### `internal/mailer/`
Mail abstraction। Development-এ terminal log, production-এ SMTP।

### `internal/httpx/`
Consistent JSON response/error format এবং safe JSON decoder।

### `internal/apperror/`
Application error code/status/message। Raw PostgreSQL বা internal error client-এর কাছে পাঠানো হয় না।

## Dependency direction

```text
handler -> service -> repository
             |          |
             |          -> PostgreSQL
             -> auth / storage / mailer
```

Repository handler-কে call করবে না। Service `http.ResponseWriter` জানবে না। Handler SQL জানবে না।

## কেন giant `utils/` নেই

`utils` folder কিছুদিন পরে unrelated code dump হয়ে যায়। তাই concern অনুযায়ী package:

- token/password -> `security`, `auth`
- email -> `mailer`
- file -> `storage`
- response -> `httpx`
- pagination -> `query`

## Production boundary

এই boilerplate single-instance/VPS deployment-এর জন্য production baseline। Multiple API instance চালালে in-memory rate limiter-এর বদলে Redis-backed distributed rate limiter ব্যবহার করবেন। SMTP/Cloudinary secrets secret manager/environment থেকে দেবেন।


## Current user status on protected requests

JWT verifies identity and expiry, then auth middleware loads the current user from PostgreSQL. `is_active` and current database role are authoritative. This means an admin role/status change applies on the next protected request instead of waiting for the access token to expire.

## Session invalidation / token version

`users.token_version` access JWT এবং refresh-token record-এর সাথে bind করা। Password change/reset অথবা logout-all token version বাড়ায়। Auth middleware PostgreSQL-এর current version check করে, তাই previously issued access token immediately invalid হয়; refresh rotation-ও user row lock নিয়ে version match করে।
