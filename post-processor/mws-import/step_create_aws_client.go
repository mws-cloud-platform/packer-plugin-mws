// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	mwsexport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export"
)

type StepCreateAWSClient struct {
	Endpoint string
	Region   string
	Path     string
}

func (s *StepCreateAWSClient) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)

	ui.Say("Creating Object Storage client...")

	hmacAccessKey := state.Get(mwsexport.HMACAccessKeyStateKey).(string)
	hmacSecretKey := state.Get(mwsexport.HMACSecretKeyStateKey).(string)
	creds := credentials.NewStaticCredentialsProvider(hmacAccessKey, hmacSecretKey, "")

	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(s.Region),
		config.WithCredentialsProvider(creds),
		config.WithBaseEndpoint(s.Endpoint),
	)
	if err != nil {
		return mws.ActionHaltWithErrorf(state, "load AWS config: %w", err)
	}

	awsClient := s3.NewPresignClient(s3.NewFromConfig(awsConfig))
	state.Put(AWSClientKey, awsClient)

	return multistep.ActionContinue
}

func (*StepCreateAWSClient) Cleanup(multistep.StateBag) {}
