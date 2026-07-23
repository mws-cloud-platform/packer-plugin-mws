// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport

import (
	"context"
	"strings"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -destination=mock/aws_mock.go . AWSClient

type AWSClient interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type StepCreateSignedLink struct {
	Path string
}

func (s *StepCreateSignedLink) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)
	awsClient := state.Get(AWSClientKey).(AWSClient)

	ui.Say("Creating presigned URL for Object Storage object...")

	bucket, key, found := strings.Cut(s.Path, "/")
	if !found {
		return mws.ActionHaltWithErrorf(state, "split object_storage_path into bucket and key: %s", s.Path)
	}
	presignResult, err := awsClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(time.Hour))
	if err != nil {
		return mws.ActionHaltWithErrorf(state, "create presigned URL: %w", err)
	}

	ui.Sayf("Presigned URL created: %s", presignResult.URL)

	state.Put(ExternalURLKey, presignResult.URL)

	return multistep.ActionContinue
}

func (*StepCreateSignedLink) Cleanup(multistep.StateBag) {}
