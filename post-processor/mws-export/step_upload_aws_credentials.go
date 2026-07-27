// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
)

const awsSharedCredsFile = "/tmp/aws_credentials" //nolint:gosec // no hardcoded credentials, only path

type StepUploadAWSSharedCredsFile struct {
}

func (s *StepUploadAWSSharedCredsFile) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(common.UIKey).(packer.Ui)
	comm := state.Get(common.CommunicatorKey).(packer.Communicator)
	hmacAccessKey := state.Get(common.HMACAccessKeyStateKey).(string)
	hmacSecretKey := state.Get(common.HMACSecretKeyStateKey).(string)

	ui.Say("Uploading AWS shared credentials file...")
	creds := fmt.Sprintf(
		"[default]\naws_access_key_id = %s\naws_secret_access_key = %s\n",
		hmacAccessKey, hmacSecretKey,
	)
	if err := comm.Upload(awsSharedCredsFile, strings.NewReader(creds), nil); err != nil {
		return common.ActionHaltWithError(state, err)
	}

	ui.Say("AWS shared credentials file uploaded")
	return multistep.ActionContinue
}

func (s *StepUploadAWSSharedCredsFile) Cleanup(multistep.StateBag) {}
