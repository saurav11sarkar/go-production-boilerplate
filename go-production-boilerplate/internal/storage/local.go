package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/saurav11sarkar/go-production-boilerplate/internal/security"
)

type Local struct {
	dir       string
	publicURL string
}

func NewLocal(dir, publicURL string) *Local {
	return &Local{dir: dir, publicURL: strings.TrimRight(publicURL, "/")}
}

func (l *Local) UploadImage(ctx context.Context, file io.Reader, filename, contentType string) (*UploadResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(l.dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload directory: %w", err)
	}
	random, err := security.RandomToken(12)
	if err != nil {
		return nil, err
	}
	ext := extensionForContentType(contentType)
	name := random + ext
	path := filepath.Join(l.dir, name)

	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return nil, fmt.Errorf("create local upload: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		_ = os.Remove(path)
		return nil, fmt.Errorf("save local upload: %w", err)
	}
	return &UploadResult{
		URL:      l.publicURL + "/uploads/" + name,
		PublicID: name,
	}, nil
}

func (l *Local) Delete(ctx context.Context, publicID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	clean := filepath.Base(publicID)
	if clean == "." || clean == "" {
		return nil
	}
	err := os.Remove(filepath.Join(l.dir, clean))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func extensionForContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
