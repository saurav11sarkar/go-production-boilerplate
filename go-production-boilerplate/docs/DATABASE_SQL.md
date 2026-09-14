# database/sql + PostgreSQL Guide

এই project application code-এ Go standard library-এর `database/sql` ব্যবহার করে। PostgreSQL connection-এর driver হলো `pgx` stdlib adapter।

## 1. Driver registration

```go
import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
)
```

Blank import (`_`) driver-কে register করে। এরপর:

```go
db, err := sql.Open("pgx", databaseURL)
```

`sql.Open` সাধারণত সঙ্গে সঙ্গে real connection verify করে না। তাই startup-এ `PingContext` করা হয়:

```go
if err := db.PingContext(ctx); err != nil {
    return err
}
```

## 2. `*sql.DB` connection না, pool

`*sql.DB` একটি concurrency-safe database handle + connection pool। এক request-এর জন্য নতুন `sql.Open` করবেন না। App startup-এ একবার open করে repository-গুলোতে একই `*sql.DB` inject করা হয়।

```go
db.SetMaxOpenConns(20)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(time.Hour)
db.SetConnMaxIdleTime(30 * time.Minute)
```

- `SetMaxOpenConns`: একই সময়ে সর্বোচ্চ open connection।
- `SetMaxIdleConns`: idle অবস্থায় pool কত connection রাখতে পারবে।
- `SetConnMaxLifetime`: একটি connection সর্বোচ্চ কতক্ষণ reuse হবে।
- `SetConnMaxIdleTime`: idle connection কতক্ষণ পরে close হবে।

Production values database capacity, app instance count এবং workload দেখে tune করবেন।

## 3. Single-row query

একটি user আনতে:

```go
row := db.QueryRowContext(ctx, `
    SELECT id, email, name
    FROM users
    WHERE id = $1
`, userID)

var user User
err := row.Scan(&user.ID, &user.Email, &user.Name)
if errors.Is(err, sql.ErrNoRows) {
    return ErrNotFound
}
if err != nil {
    return err
}
```

PostgreSQL placeholders `$1`, `$2`, `$3` ব্যবহার করুন; string concatenate করে SQL বানাবেন না।

## 4. Multiple-row query

```go
rows, err := db.QueryContext(ctx, `
    SELECT id, email, name
    FROM users
    ORDER BY created_at DESC
    LIMIT $1 OFFSET $2
`, limit, offset)
if err != nil {
    return err
}
defer rows.Close()

for rows.Next() {
    var user User
    if err := rows.Scan(&user.ID, &user.Email, &user.Name); err != nil {
        return err
    }
}

if err := rows.Err(); err != nil {
    return err
}
```

`rows.Close()` এবং শেষে `rows.Err()` check করা production code-এ important।

## 5. INSERT/UPDATE/DELETE without returned row

```go
result, err := db.ExecContext(ctx, `
    UPDATE users
    SET is_active = $2
    WHERE id = $1
`, userID, active)
if err != nil {
    return err
}

count, err := result.RowsAffected()
if err != nil {
    return err
}
if count == 0 {
    return ErrNotFound
}
```

কিন্তু PostgreSQL `RETURNING` দিয়ে value ফেরত দরকার হলে `QueryRowContext` use করবেন।

## 6. Transaction

যখন কয়েকটি write একসাথে success/fail হওয়া দরকার:

```go
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return err
}
defer tx.Rollback()

if _, err := tx.ExecContext(ctx, query1, arg1); err != nil {
    return err
}

if _, err := tx.ExecContext(ctx, query2, arg2); err != nil {
    return err
}

if err := tx.Commit(); err != nil {
    return err
}
```

এই project-এর `database.WithTx` helper একই pattern reusable করেছে।

## 7. Row lock

Refresh token rotation-এর মতো concurrency-sensitive flow-এ:

```sql
SELECT ...
FROM refresh_tokens
WHERE token_hash = $1
FOR UPDATE;
```

এটা transaction-এর মধ্যে selected row lock করে। Concurrent request একই token একসাথে rotate করতে পারবে না।

## 8. Context কেন

প্রায় সব database operation-এ `Context` variant use করুন:

```text
QueryRowContext
QueryContext
ExecContext
PingContext
BeginTx
```

HTTP client disconnect/timeout হলে context cancellation database operation পর্যন্ত propagate হতে পারে।

## 9. Driver বদলালে কী হবে?

`database/sql` concepts একই থাকবে, কিন্তু project-এর SQL PostgreSQL-specific। যেমন:

```text
PostgreSQL placeholder: $1
MySQL placeholder:      ?
PostgreSQL search:      ILIKE
PostgreSQL UUID/enum এবং locking syntax-ও আলাদা হতে পারে
```

তাই MySQL driver লাগালেই existing PostgreSQL query 100% unchanged চলবে—এমন ধরে নেবেন না। তবে `*sql.DB`, context, transactions, repository pattern এবং testing concepts reusable থাকবে।

## 10. এই project-এর flow

```text
cmd/api
  -> database.Connect()
      -> sql.Open("pgx", DATABASE_URL)
      -> pool configuration
      -> PingContext
  -> repository.NewUserRepository(db)
  -> service
  -> handler
```

Repository database handle close করে না। App shutdown-এর সময় `main` একবার `db.Close()` করে।
