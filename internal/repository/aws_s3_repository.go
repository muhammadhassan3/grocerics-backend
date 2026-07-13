package repository

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"grocerics-backend/internal/config"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSRepository struct {
	client *config.AWSClient
}

func NewAWSRepository(client *config.AWSClient) *AWSRepository {
	return &AWSRepository{
		client: client,
	}
}

type PresignedURLResponse struct {
	URL       string
	PublicURL string
	ExipresAt time.Time
}

func (r *AWSRepository) GetPresignedURL(folder, filename, contentType string) (*PresignedURLResponse, error) {
	if r.client == nil {
		return nil, errors.New("AWS client is not initialized")
	}
	if folder == "" || filename == "" || contentType == "" {
		return nil, errors.New("folder, filename, and contentType must be provided")
	}

	if !checkValidContentType(contentType) {
		return nil, errors.New("invalid content type")
	}

	key := generateObjectKey(folder, filename)
	expiry := time.Duration(15 * time.Minute)

	request, err := r.client.PresignClient.PresignPutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      &r.client.BucketName,
		Key:         &key,
		ContentType: &contentType,
	}, func(opt *s3.PresignOptions) {
		opt.Expires = expiry
	})
	if err != nil {
		return nil, err
	}

	publicURL := generatePublicURL(r.client.BucketName, key, r.client.Region)

	return &PresignedURLResponse{
		URL:       request.URL,
		PublicURL: publicURL,
		ExipresAt: time.Now().Add(expiry),
	}, nil
}

func generatePublicURL(bucketName, key, region string) string {
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, region, key)
}

func generateObjectKey(folder, filename string) string {
	// if filename is already sanitized, we can just use it as is
	if isSanitized(filename) {
		return folder + "/" + filename
	}

	ext := filepath.Ext(filename)
	baseName := strings.TrimSuffix(filename, ext)

	safeName := strings.ReplaceAll(baseName, " ", "_")

	if len(safeName) > 50 {
		safeName = safeName[:50]
	}

	dateStr := time.Now().UnixMilli()
	return fmt.Sprintf("%s/%d-%s%s", folder, dateStr, safeName, ext)
}

func isSanitized(filename string) bool {
	// Check for invalid characters in the filename
	values := strings.Split(filename, "-")
	if len(values) > 1 {
		jsTimestampPrefixPattern := regexp.MustCompile(`^(\d{13})-(.+)$`)
		matches := jsTimestampPrefixPattern.FindStringSubmatch(filename)

		return matches != nil
	}

	// Check for length constraints
	if len(filename) > 255 {
		return false
	}

	return true
}

func checkValidContentType(contentType string) bool {
	allowedContentTypes := []string{
		"image/jpeg",
		"image/png",
		"image/jpg",
		"image/webp",
	}

	for _, allowed := range allowedContentTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}
