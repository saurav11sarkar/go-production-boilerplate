package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	AppName     string
	PublicURL   string
	FrontendURL string

	HTTP struct {
		Port                  string
		ReadTimeout           time.Duration
		WriteTimeout          time.Duration
		IdleTimeout           time.Duration
		RequestTimeout        time.Duration
		ShutdownTimeout       time.Duration
		TrustedProxy          bool
		AllowedOrigins        []string
		RateLimitRequests     int
		RateLimitWindow       time.Duration
		AuthRateLimitRequests int
		AuthRateLimitWindow   time.Duration
	}

	Database struct {
		URL             string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
		ConnMaxIdleTime time.Duration
	}

	Auth struct {
		Issuer                   string
		AccessTokenSecret        string
		AccessTokenTTL           time.Duration
		RefreshTokenTTL          time.Duration
		RefreshCookieName        string
		RefreshCookieSecure      bool
		RefreshCookieSameSite    string
		RefreshCookieDomain      string
		RefreshTokenInBody       bool
		RequireEmailVerification bool
	}

	Storage struct {
		Provider            string
		LocalDir            string
		CloudinaryCloudName string
		CloudinaryAPIKey    string
		CloudinaryAPISecret string
		MaxImageBytes       int64
	}

	Mail struct {
		Provider string
		FromName string
		From     string
		Host     string
		Port     int
		Username string
		Password string
		Timeout  time.Duration
	}
}

func MustLoad() Config {
	_ = godotenv.Load()

	var cfg Config
	cfg.Environment = getEnv("APP_ENV", "development")
	cfg.AppName = getEnv("APP_NAME", "Go Production API")
	cfg.PublicURL = strings.TrimRight(getEnv("PUBLIC_BASE_URL", "http://localhost:8080"), "/")
	cfg.FrontendURL = strings.TrimRight(getEnv("FRONTEND_URL", "http://localhost:3000"), "/")

	cfg.HTTP.Port = getEnv("HTTP_PORT", "8080")
	cfg.HTTP.ReadTimeout = mustDuration("HTTP_READ_TIMEOUT", "15s")
	cfg.HTTP.WriteTimeout = mustDuration("HTTP_WRITE_TIMEOUT", "30s")
	cfg.HTTP.IdleTimeout = mustDuration("HTTP_IDLE_TIMEOUT", "60s")
	cfg.HTTP.RequestTimeout = mustDuration("HTTP_REQUEST_TIMEOUT", "20s")
	cfg.HTTP.ShutdownTimeout = mustDuration("HTTP_SHUTDOWN_TIMEOUT", "10s")
	cfg.HTTP.TrustedProxy = mustBool("TRUST_PROXY", false)
	cfg.HTTP.AllowedOrigins = splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"))
	cfg.HTTP.RateLimitRequests = mustInt("RATE_LIMIT_REQUESTS", 120)
	cfg.HTTP.RateLimitWindow = mustDuration("RATE_LIMIT_WINDOW", "1m")
	cfg.HTTP.AuthRateLimitRequests = mustInt("AUTH_RATE_LIMIT_REQUESTS", 10)
	cfg.HTTP.AuthRateLimitWindow = mustDuration("AUTH_RATE_LIMIT_WINDOW", "1m")

	cfg.Database.URL = mustEnv("DATABASE_URL")
	cfg.Database.MaxOpenConns = mustInt("DB_MAX_OPEN_CONNS", 20)
	cfg.Database.MaxIdleConns = mustInt("DB_MAX_IDLE_CONNS", 5)
	cfg.Database.ConnMaxLifetime = mustDuration("DB_CONN_MAX_LIFETIME", "1h")
	cfg.Database.ConnMaxIdleTime = mustDuration("DB_CONN_MAX_IDLE_TIME", "30m")

	cfg.Auth.Issuer = getEnv("JWT_ISSUER", "go-production-api")
	cfg.Auth.AccessTokenSecret = mustEnv("JWT_ACCESS_SECRET")
	if len(cfg.Auth.AccessTokenSecret) < 32 {
		panic("JWT_ACCESS_SECRET must be at least 32 characters")
	}
	cfg.Auth.AccessTokenTTL = mustDuration("ACCESS_TOKEN_TTL", "15m")
	cfg.Auth.RefreshTokenTTL = mustDuration("REFRESH_TOKEN_TTL", "720h")
	cfg.Auth.RefreshCookieName = getEnv("REFRESH_COOKIE_NAME", "refresh_token")
	cfg.Auth.RefreshCookieSecure = mustBool("REFRESH_COOKIE_SECURE", cfg.Environment == "production")
	cfg.Auth.RefreshCookieSameSite = strings.ToLower(getEnv("REFRESH_COOKIE_SAMESITE", "lax"))
	cfg.Auth.RefreshCookieDomain = strings.TrimSpace(os.Getenv("REFRESH_COOKIE_DOMAIN"))
	cfg.Auth.RefreshTokenInBody = mustBool("REFRESH_TOKEN_IN_BODY", false)
	cfg.Auth.RequireEmailVerification = mustBool("REQUIRE_EMAIL_VERIFICATION", true)

	cfg.Storage.Provider = strings.ToLower(getEnv("STORAGE_PROVIDER", "local"))
	cfg.Storage.LocalDir = getEnv("LOCAL_UPLOAD_DIR", "uploads")
	cfg.Storage.CloudinaryCloudName = os.Getenv("CLOUDINARY_CLOUD_NAME")
	cfg.Storage.CloudinaryAPIKey = os.Getenv("CLOUDINARY_API_KEY")
	cfg.Storage.CloudinaryAPISecret = os.Getenv("CLOUDINARY_API_SECRET")
	cfg.Storage.MaxImageBytes = int64(mustInt("MAX_IMAGE_MB", 5)) * 1024 * 1024

	cfg.Mail.Provider = strings.ToLower(getEnv("MAIL_PROVIDER", "log"))
	cfg.Mail.FromName = getEnv("MAIL_FROM_NAME", cfg.AppName)
	cfg.Mail.From = getEnv("MAIL_FROM", "noreply@example.com")
	cfg.Mail.Host = os.Getenv("SMTP_HOST")
	cfg.Mail.Port = mustInt("SMTP_PORT", 587)
	cfg.Mail.Username = os.Getenv("SMTP_USERNAME")
	cfg.Mail.Password = os.Getenv("SMTP_PASSWORD")
	cfg.Mail.Timeout = mustDuration("SMTP_TIMEOUT", "10s")

	if cfg.Database.MaxOpenConns < 1 || cfg.Database.MaxIdleConns < 0 || cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		panic("invalid database/sql pool sizes")
	}
	if cfg.HTTP.RateLimitRequests < 1 {
		panic("RATE_LIMIT_REQUESTS must be at least 1")
	}
	if cfg.HTTP.AuthRateLimitRequests < 1 {
		panic("AUTH_RATE_LIMIT_REQUESTS must be at least 1")
	}
	if cfg.Storage.MaxImageBytes < 1 {
		panic("MAX_IMAGE_MB must be at least 1")
	}
	if cfg.Mail.Port < 1 || cfg.Mail.Port > 65535 {
		panic("SMTP_PORT must be between 1 and 65535")
	}
	switch cfg.Auth.RefreshCookieSameSite {
	case "lax", "strict", "none":
	default:
		panic("REFRESH_COOKIE_SAMESITE must be lax, strict, or none")
	}
	if cfg.Auth.RefreshCookieSameSite == "none" && !cfg.Auth.RefreshCookieSecure {
		panic("REFRESH_COOKIE_SECURE must be true when REFRESH_COOKIE_SAMESITE=none")
	}
	if cfg.Environment == "production" {
		if !cfg.Auth.RefreshCookieSecure {
			panic("REFRESH_COOKIE_SECURE must be true in production")
		}
		if strings.Contains(cfg.Auth.AccessTokenSecret, "replace-this") {
			panic("JWT_ACCESS_SECRET must be replaced in production")
		}
		if cfg.Mail.Provider == "log" {
			panic("MAIL_PROVIDER=log is not allowed in production because auth tokens would be written to logs")
		}
		if !strings.HasPrefix(cfg.PublicURL, "https://") || !strings.HasPrefix(cfg.FrontendURL, "https://") {
			panic("PUBLIC_BASE_URL and FRONTEND_URL must use https:// in production")
		}
	}

	return cfg
}

func mustEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		panic(fmt.Sprintf("%s is required", key))
	}
	return value
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func mustDuration(key, fallback string) time.Duration {
	value := getEnv(key, fallback)
	d, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: %v", key, err))
	}
	return d
}

func mustInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: %v", key, err))
	}
	return n
}

func mustBool(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: %v", key, err))
	}
	return b
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
