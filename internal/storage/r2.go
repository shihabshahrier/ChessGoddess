package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2Client struct {
	client *s3.Client
	bucket string
	uploader *manager.Uploader
}

func NewR2Client(accessKey, secretKey, bucket, endpoint string) (*R2Client, error) {
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("R2 credentials not configured")
	}

	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")

	cfg := aws.Config{
		Region:       "auto",
		Credentials:  creds,
		BaseEndpoint: aws.String(endpoint),
	}

	client := s3.NewFromConfig(cfg)

	return &R2Client{
		client:   client,
		bucket:   bucket,
		uploader: manager.NewUploader(client),
	}, nil
}

func (c *R2Client) UploadImage(ctx context.Context, key string, reader io.Reader, contentType string) (string, error) {
	_, err := c.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	return key, nil
}

func (s *R2Client) UploadScreenshot(ctx context.Context, userID string, data []byte) (string, error) {
	key := fmt.Sprintf("screenshots/%s/%s.png", userID, generateKey())
	return s.UploadImage(ctx, key, bytes.NewReader(data), "image/png")
}

func (c *R2Client) UploadThumbnail(ctx context.Context, userID string, data []byte) (string, error) {
	key := fmt.Sprintf("thumbnails/%s/%s.jpg", userID, generateKey())
	return c.UploadImage(ctx, key, bytes.NewReader(data), "image/jpeg")
}

func (c *R2Client) GetURL(key string) string {
	endpoint := ""
	if c.client.Options().BaseEndpoint != nil {
		endpoint = *c.client.Options().BaseEndpoint
	}
	return fmt.Sprintf("https://%s.%s/%s", c.bucket, endpoint, key)
}

func (c *R2Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete from R2: %w", err)
	}

	return nil
}

func generateKey() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
