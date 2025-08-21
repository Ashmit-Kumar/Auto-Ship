// internal/services/s3.go
package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var s3Client *s3.Client
var s3Bucket string
var s3WebsiteURL string // base URL for static hosting

func InitS3() error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("unable to load AWS config: %v", err)
	}

	s3Client = s3.NewFromConfig(cfg)

	s3Bucket = os.Getenv("S3_BUCKET_NAME")
	if s3Bucket == "" {
		return fmt.Errorf("S3_BUCKET_NAME not set")
	}

	region := os.Getenv("AWS_REGION")
	if region == "" {
		return fmt.Errorf("AWS_REGION not set")
	}

	s3WebsiteURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", s3Bucket, region)

	return nil
}

// UploadStaticSite uploads a folder recursively to S3 (uncompressed).
func UploadStaticSite(localPath, keyPrefix string) (string, error) {
	if s3Client == nil {
		return "", fmt.Errorf("S3 client not initialized")
	}

	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}

		key := filepath.ToSlash(filepath.Join(keyPrefix, relPath)) // Use forward slashes
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(s3Bucket),
			Key:    aws.String(key),
			Body:   file,
			// S3 Website hosting doesn't require ACL if bucket is public or via CloudFront
			ContentType: aws.String(getContentTypeByExtension(path)),
		})
		if err != nil {
			return fmt.Errorf("failed to upload file %s: %v", key, err)
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	// Return URL to index.html
	indexURL := fmt.Sprintf("%s/%s/index.html", s3WebsiteURL, keyPrefix)
	// indexURL := fmt.Sprintf("%s/%s/", s3WebsiteURL, keyPrefix)

	return indexURL, nil
}

// getContentTypeByExtension returns basic content-type based on file extension.
func getContentTypeByExtension(filePath string) string {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	default:
		return "application/octet-stream"
	}
}
