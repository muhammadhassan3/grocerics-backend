package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AWSClient struct {
	Client        *s3.Client
	PresignClient *s3.PresignClient
	BucketName    string
	Region        string
}

func NewAWSClient(cfg AWSConfig) (*AWSClient, error) {
	config, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(cfg.Region), config.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
	))
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(config)
	presignClient := s3.NewPresignClient(client)
	return &AWSClient{
		Client:        client,
		PresignClient: presignClient,
		BucketName:    cfg.BucketName,
		Region:        cfg.Region,
	}, nil
}
