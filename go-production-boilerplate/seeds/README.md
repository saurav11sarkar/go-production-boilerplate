# Seeds

Use the Go seed command instead of storing a plaintext admin password in SQL:

```bash
go run ./cmd/seed
```

Required environment variables:

- `SEED_ADMIN_NAME`
- `SEED_ADMIN_EMAIL`
- `SEED_ADMIN_PASSWORD`

The command hashes the password with bcrypt and safely upserts the admin user.
