// internal/services/s3.go
package services

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var s3Client *s3.Client
var s3Bucket string

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
	return nil
}

func zipFolder(source string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(source, path)
		if info.IsDir() {
			if relPath != "." {
				_, err := zipWriter.Create(relPath + "/")
				return err
			}
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		w, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, file)
		return err
	})

	if err != nil {
		return nil, err
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

func UploadToS3(folderPath, key string) (string, error) {
	zipBuf, err := zipFolder(folderPath)
	if err != nil {
		return "", fmt.Errorf("failed to zip folder: %v", err)
	}

	_, err = s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s3Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(zipBuf.Bytes()),
		ContentType: aws.String("application/zip"),
		ACL:         s3types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	publicURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s3Bucket, key)
	return publicURL, nil
}
