package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/config"
)

type UploadResult struct {
	URL      string
	PublicID string
}

type Storage interface {
	UploadImage(ctx context.Context, file io.Reader, filename, contentType string) (*UploadResult, error)
	Delete(ctx context.Context, publicID string) error
}

func New(cfg config.Config) (Storage, error) {
	switch cfg.Storage.Provider {
	case "local":
		return NewLocal(cfg.Storage.LocalDir, cfg.PublicURL), nil
	case "cloudinary":
		if cfg.Storage.CloudinaryCloudName == "" || cfg.Storage.CloudinaryAPIKey == "" || cfg.Storage.CloudinaryAPISecret == "" {
			return nil, fmt.Errorf("cloudinary credentials are required when STORAGE_PROVIDER=cloudinary")
		}
		return NewCloudinary(cfg.Storage.CloudinaryCloudName, cfg.Storage.CloudinaryAPIKey, cfg.Storage.CloudinaryAPISecret)
	default:
		return nil, fmt.Errorf("unsupported storage provider %q", cfg.Storage.Provider)
	}
}
