package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type Cloudinary struct {
	client *cloudinary.Cloudinary
}

func NewCloudinary(cloudName, apiKey, apiSecret string) (*Cloudinary, error) {
	client, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("create cloudinary client: %w", err)
	}
	return &Cloudinary{client: client}, nil
}

func (c *Cloudinary) UploadImage(ctx context.Context, file io.Reader, filename, contentType string) (*UploadResult, error) {
	result, err := c.client.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:       "go-production-api/avatars",
		ResourceType: "image",
	})
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload: %w", err)
	}
	return &UploadResult{URL: result.SecureURL, PublicID: result.PublicID}, nil
}

func (c *Cloudinary) Delete(ctx context.Context, publicID string) error {
	if publicID == "" {
		return nil
	}
	_, err := c.client.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:     publicID,
		ResourceType: "image",
	})
	if err != nil {
		return fmt.Errorf("cloudinary delete: %w", err)
	}
	return nil
}
