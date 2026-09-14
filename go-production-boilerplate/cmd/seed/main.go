package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/config"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/database"
	"github.com/saurav11sarkar/go-production-boilerplate/internal/security"
)

func main() {
	cfg := config.MustLoad()
	email := strings.ToLower(strings.TrimSpace(os.Getenv("SEED_ADMIN_EMAIL")))
	password := os.Getenv("SEED_ADMIN_PASSWORD")
	name := strings.TrimSpace(os.Getenv("SEED_ADMIN_NAME"))
	if name == "" {
		name = "Admin"
	}
	if email == "" || password == "" {
		panic("SEED_ADMIN_EMAIL and SEED_ADMIN_PASSWORD are required")
	}
	if len(password) < 8 {
		panic("SEED_ADMIN_PASSWORD must be at least 8 characters")
	}

	ctx := context.Background()
	db, err := database.Connect(ctx, cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	hash, err := security.NewPasswordManager().Hash(password)
	if err != nil {
		panic(err)
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (name, email, password_hash, role, is_verified, is_active)
		VALUES ($1, LOWER($2), $3, 'admin', TRUE, TRUE)
		ON CONFLICT (email) DO UPDATE
		SET name = EXCLUDED.name,
		    password_hash = EXCLUDED.password_hash,
		    role = 'admin',
		    is_verified = TRUE,
		    is_active = TRUE,
		    token_version = users.token_version + 1,
		    updated_at = NOW()`, name, email, hash)
	if err != nil {
		panic(err)
	}
	fmt.Printf("admin seed complete: %s\n", email)
}
