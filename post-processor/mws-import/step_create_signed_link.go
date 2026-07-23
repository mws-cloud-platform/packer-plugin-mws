// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	mwsexport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export"
)

type StepCreateSignedLink struct {
	Endpoint string
	Region   string
	Path     string
}

func (s *StepCreateSignedLink) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)

	ui.Say("Creating presigned URL for Object Storage object...")

	hmacAccessKey := state.Get(mwsexport.HMACAccessKeyStateKey).(string)
	hmacSecretKey := state.Get(mwsexport.HMACSecretKeyStateKey).(string)
	creds := credentials.NewStaticCredentialsProvider(hmacAccessKey, hmacSecretKey, "")

	s3Config, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s.Region),
		config.WithCredentialsProvider(creds),
		config.WithBaseEndpoint(s.Endpoint),
	)
	if err != nil {
		return mws.ActionHaltWithErrorf(state, "failed to load AWS config: %w", err)
	}

	s3Client := s3.NewFromConfig(s3Config)

	presignClient := s3.NewPresignClient(s3Client)

	bucket, key, found := strings.Cut(s.Path, "/")
	if !found {
		return mws.ActionHaltWithErrorf(state, "failed to split object_storage_path into bucket and key: %s", s.Path)
	}
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return mws.ActionHaltWithErrorf(state, "failed to create presigned URL: %w", err)
	}

	ui.Sayf("Presigned URL created: %s", presignResult.URL)

	state.Put(ExternalURLKey, presignResult.URL)

	return multistep.ActionContinue
}

func (*StepCreateSignedLink) Cleanup(multistep.StateBag) {}
