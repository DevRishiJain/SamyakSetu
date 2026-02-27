package services

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3StorageService implements StorageService using AWS S3.
type S3StorageService struct {
	client     *s3.Client
	bucketName string
	region     string
}

// NewS3StorageService initializes a new S3 client and returns the S3StorageService.
func NewS3StorageService(region, accessKey, secretKey, bucketName string) (*S3StorageService, error) {
	// Create custom credentials provider
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(creds),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	log.Printf("INFO: AWS S3 storage initialized to bucket: %s", bucketName)

	return &S3StorageService{
		client:     client,
		bucketName: bucketName,
		region:     region,
	}, nil
}

// SaveFile uploads a file to S3 and returns the public URL.
func (s *S3StorageService) SaveFile(file *multipart.FileHeader, subDir string) (string, error) {
	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	uniqueName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	// Create the S3 key (path inside bucket)
	// Example: soil/1709230501.jpg
	s3Key := fmt.Sprintf("%s/%s", subDir, uniqueName)

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Get file content type
	contentType := file.Header.Get("Content-Type")

	// Upload directly to S3
	_, err = s.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(s3Key),
		Body:        src,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Construct public URL
	publicURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, s3Key)
	return publicURL, nil
}
