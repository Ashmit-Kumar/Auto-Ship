package cloud

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	awsv1 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// awsProvider hosts static sites on S3 and opens container ports in an EC2
// security group. It preserves the behavior of the original services.InitS3 /
// services.UploadStaticSite / utils.AuthorizeEC2Port code.
type awsProvider struct {
	s3Client   *s3.Client
	bucket     string
	websiteURL string // base URL for static hosting
	region     string
	sgID       string // EC2 security group id (only needed for dynamic projects)
}

// newAWSProvider loads AWS config and validates the env vars required for static
// hosting. The security-group id is optional here and only enforced when a
// dynamic project actually needs a port opened (see AuthorizePort).
func newAWSProvider() (Provider, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	bucket := os.Getenv("S3_BUCKET_NAME")
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME not set")
	}
	region := os.Getenv("AWS_REGION")
	if region == "" {
		return nil, fmt.Errorf("AWS_REGION not set")
	}

	return &awsProvider{
		s3Client:   s3.NewFromConfig(cfg),
		bucket:     bucket,
		websiteURL: fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region),
		region:     region,
		sgID:       os.Getenv("EC2_SECURITY_GROUP_ID"),
	}, nil
}

func (a *awsProvider) Name() string { return "aws" }

// UploadStaticSite uploads a folder recursively to S3 (uncompressed) and returns
// the URL to its index.html.
func (a *awsProvider) UploadStaticSite(localPath, keyPrefix string) (string, error) {
	err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localPath, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(filepath.Join(keyPrefix, relPath))

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = a.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket:      awsv2.String(a.bucket),
			Key:         awsv2.String(key),
			Body:        file,
			ContentType: awsv2.String(contentTypeByExtension(path)),
		})
		if err != nil {
			return fmt.Errorf("failed to upload file %s: %w", key, err)
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s/index.html", a.websiteURL, keyPrefix), nil
}

// AuthorizePort opens an inbound TCP rule for port in the configured EC2 security
// group (0.0.0.0/0). A duplicate rule is treated as success.
func (a *awsProvider) AuthorizePort(port int) error {
	if a.sgID == "" {
		return fmt.Errorf("EC2_SECURITY_GROUP_ID not set; cannot open port %d", port)
	}

	sess := session.Must(session.NewSession(&awsv1.Config{
		Region: awsv1.String(a.region),
	}))
	svc := ec2.New(sess)

	_, err := svc.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
		GroupId: awsv1.String(a.sgID),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: awsv1.String("tcp"),
				FromPort:   awsv1.Int64(int64(port)),
				ToPort:     awsv1.Int64(int64(port)),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      awsv1.String("0.0.0.0/0"),
						Description: awsv1.String("Auto-opened for container hosting"),
					},
				},
			},
		},
	})
	if err != nil && !strings.Contains(err.Error(), "InvalidPermission.Duplicate") {
		return err
	}
	return nil
}
