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
	s3PullBucket  string
	s3PullKey     string
	s3PullTimeout int
)

// adbS3PullCmd represents the s3-pull command
var adbS3PullCmd = &cobra.Command{
	Use:   "s3-pull <cache-file>",
	Args:  cobra.ExactArgs(1),
	Short: "Download cache from S3 before processing",
	Long: `This downloads the AbuseIPDb cache file from S3 to a local path.

The S3 bucket can be specified via --bucket flag or AWS_BUCKET environment variable.
S3 credentials must be set in AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.`,
	Run: func(_ *cobra.Command, args []string) {
		localPath := args[0]

		// Resolve bucket from flag or environment variable
		bucket := s3PullBucket
		if bucket == "" {
			bucket = os.Getenv("AWS_BUCKET")
		}
		if bucket == "" {
			fmt.Fprintln(os.Stderr, "S3 bucket must be specified via --bucket flag or AWS_BUCKET environment variable")
			os.Exit(1)
		}

		// Resolve key from flag or use cache file basename
		key := s3PullKey
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
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s3PullTimeout)*time.Second)
		defer cancel()

		// Download from S3
		if err := client.Download(ctx, bucket, key, localPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading from S3: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully downloaded %s from s3://%s/%s\n", localPath, bucket, key)
	},
}

func init() {
	adbCmd.AddCommand(adbS3PullCmd)
	adbS3PullCmd.Flags().StringVar(&s3PullBucket, "bucket", "", "S3 bucket (or AWS_BUCKET environment variable)")
	adbS3PullCmd.Flags().StringVar(&s3PullKey, "key", "", "S3 object key (default: cache-file name)")
	adbS3PullCmd.Flags().IntVar(&s3PullTimeout, "timeout", 30, "Timeout in seconds")
}
