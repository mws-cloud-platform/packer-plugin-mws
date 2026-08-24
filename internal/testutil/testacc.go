// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

// Package testutil provides helper functions for testing.
package testutil

import (
	"context"
	"fmt"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
)

func LoadAWSClient(ctx context.Context, accessKey, secretKey string) (*s3.Client, error) {
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(config.DefaultObjectStorageRegion),
		awsconfig.WithCredentialsProvider(creds),
		awsconfig.WithBaseEndpoint(config.DefaultObjectStorageEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	awsClient := s3.NewFromConfig(awsConfig)

	return awsClient, nil
}
