package cmd

import (
	"context"
	"fmt"
	"iplists/cmd/internal/s3"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

var (
	s3PushBucket  string
	s3PushKey     string
	s3PushTimeout int
)

// adbS3PushCmd represents the s3-push command
var adbS3PushCmd = &cobra.Command{
	Use:   "s3-push <cache-file>",
	Args:  cobra.ExactArgs(1),
	Short: "Upload cache to S3 after processing",
	Long: `This uploads the AbuseIPDb cache file to S3 from a local path.

The S3 bucket can be specified via --bucket flag or AWS_BUCKET environment variable.
S3 credentials must be set in AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.`,
	Run: func(_ *cobra.Command, args []string) {
		localPath := args[0]

		// Verify cache file exists
		if _, err := os.Stat(localPath); err != nil {
			fmt.Fprintf(os.Stderr, "Cache file not found at %s: %v\n", localPath, err)
			os.Exit(1)
		}

		// Resolve bucket from flag or environment variable
		bucket := s3PushBucket
		if bucket == "" {
			bucket = os.Getenv("AWS_BUCKET")
		}
		if bucket == "" {
			fmt.Fprintln(os.Stderr, "S3 bucket must be specified via --bucket flag or AWS_BUCKET environment variable")
			os.Exit(1)
		}

		// Resolve key from flag or use cache file basename
		key := s3PushKey
		if key == "" {
			key = filepath.Base(localPath)
		}
		// Create S3 client from environment
		client, err := s3.NewClientFromEnv()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing S3 client: %v\n", err)
			os.Exit(1)
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s3PushTimeout)*time.Second)
		defer cancel()

		// Upload to S3
		if err := client.Upload(ctx, bucket, key, localPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error uploading to S3: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully uploaded %s to s3://%s/%s\n", localPath, bucket, key)
	},
}

func init() {
	adbCmd.AddCommand(adbS3PushCmd)
	adbS3PushCmd.Flags().StringVar(&s3PushBucket, "bucket", "", "S3 bucket (or AWS_BUCKET environment variable)")
	adbS3PushCmd.Flags().StringVar(&s3PushKey, "key", "", "S3 object key (default: cache-file name)")
	adbS3PushCmd.Flags().IntVar(&s3PushTimeout, "timeout", 30, "Timeout in seconds")
}
