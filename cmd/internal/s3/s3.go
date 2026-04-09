// Package s3 provides S3 operations for cache management.
package s3

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps S3 operations with connection management.
type Client struct {
	minioClient *minio.Client
	region      string
}

// NewClientFromEnv creates an S3 client from environment variables.
// Required env vars: AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
// Optional env vars: AWS_ENDPOINT, AWS_REGION, AWS_USE_PATH_STYLE
// Returns error if required credentials are missing.
func NewClientFromEnv() (*Client, error) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	endpoint := os.Getenv("AWS_ENDPOINT")

	if accessKeyID == "" || secretAccessKey == "" || endpoint == "" {
		return nil, fmt.Errorf("AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_ENDPOINT environment variables must be set")
	}

	region := os.Getenv("AWS_REGION")

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
		Region: region,
	}

	minioClient, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize S3 client: %w", err)
	}

	return &Client{
		minioClient: minioClient,
		region:      region,
	}, nil
}

// Download retrieves a compressed object from S3 and saves as uncompressed local file.
// The key is automatically suffixed with ".gz" to locate the remote object.
// Returns error if bucket/key not found or write fails.
func (c *Client) Download(ctx context.Context, bucket, key, localPath string) error {
	object, err := c.minioClient.GetObject(ctx, bucket, key+".gz", minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object %s from bucket %s: %w", key, bucket, err)
	}
	defer func() { _ = object.Close() }()

	// Verify object exists
	if _, err := object.Stat(); err != nil {
		return fmt.Errorf("failed to stat object %s/%s.gz: %w", bucket, key, err)
	}

	// Create gzip reader to decompress the S3 object
	gzipReader, err := gzip.NewReader(object)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader for %s.gz: %w", key, err)
	}
	defer func() { _ = gzipReader.Close() }()

	// Create or truncate local file
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer func() { _ = file.Close() }()

	// Write decompressed data to local file
	if _, err := io.Copy(file, gzipReader); err != nil {
		return fmt.Errorf("failed to write to local file %s: %w", localPath, err)
	}

	return nil
}

// Upload compresses a local file and uploads it to S3 with a ".gz" suffix appended to key.
// Returns error if file not found or upload fails.
func (c *Client) Upload(ctx context.Context, bucket, key, localPath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %w", localPath, err)
	}
	defer func() { _ = file.Close() }()

	// Create a pipe for streaming compressed data
	reader, writer := io.Pipe()
	defer func() { _ = reader.Close() }()

	// Variable to store compression error from goroutine
	var compressErr error
	go func() {
		defer func() { _ = writer.Close() }()

		gzipWriter := gzip.NewWriter(writer)
		if _, err := io.Copy(gzipWriter, file); err != nil {
			compressErr = err
		}
		if err := gzipWriter.Close(); err != nil && compressErr == nil {
			compressErr = err
		}
	}()

	// Upload compressed data from the pipe
	_, err = c.minioClient.PutObject(
		ctx,
		bucket,
		key+".gz",
		reader,
		-1, // Unknown size due to streaming compression
		minio.PutObjectOptions{
			ContentType: "application/gzip",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload to bucket %s, key %s: %w", bucket, key, err)
	}

	// Check if compression had any errors
	if compressErr != nil {
		return fmt.Errorf("error during compression of %s: %w", localPath, compressErr)
	}

	return nil
}
