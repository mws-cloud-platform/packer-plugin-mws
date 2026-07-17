// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
)

type StepUploadImage struct {
	Endpoint string
	Region   string
	Path     string
}

func (s *StepUploadImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)
	comm := state.Get(mws.CommunicatorKey).(packer.Communicator)

	uploadCmd := &packer.RemoteCmd{
		Command: fmt.Sprintf(
			"%s=%s aws s3 --region=%s --endpoint-url=%s cp %s s3://%s",
			"AWS_SHARED_CREDENTIALS_FILE",
			awsSharedCredsFile,
			s.Region,
			s.Endpoint,
			diskImageFile,
			s.Path,
		),
	}

	ui.Sayf("Uploading image to %q...", s.Path)
	if err := uploadCmd.RunWithUi(ctx, comm, ui); err != nil {
		return mws.ActionHaltWithError(state, err)
	}
	if code := uploadCmd.ExitStatus(); code != 0 {
		return mws.ActionHaltWithErrorf(state, "bad exit code: %d", code)
	}

	ui.Say("Image uploaded")
	return multistep.ActionContinue
}

func (s *StepUploadImage) Cleanup(multistep.StateBag) {}
