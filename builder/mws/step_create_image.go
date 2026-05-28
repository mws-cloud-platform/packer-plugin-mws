// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
)

type StepCreateImage struct {
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(configKey).(*Config)
	driver := state.Get(driverKey).(Driver)
	prefix := state.Get(uuidPrefixKey).(string)

	ui := state.Get(uiKey).(packer.Ui)

	ui.Say("Creating image from virtual machine...")

	diskRef, ok := state.Get(diskRefKey).(*computeref.DiskRef)
	if !ok || diskRef == nil {
		return actionHaltWithError(state, fmt.Errorf("disk ref not found in state: %w", errUnexpected))
	}

	imageName := config.ImageName
	if imageName == "" {
		imageName = prefix + "image"
	}

	image, err := driver.CreateImage(ctx, CreateImageParams{
		ImageName:        imageName,
		ImageDescription: config.ImageDescription,
		DiskRef:          diskRef,
	})
	if err != nil {
		return actionHaltWithError(state, fmt.Errorf("create image: %w", err))
	}

	ui.Sayf("Created image %q", imageName)

	state.Put(imageKey, image)

	s.GeneratedData.Put("ImageName", imageName)

	return multistep.ActionContinue
}

func (*StepCreateImage) Cleanup(multistep.StateBag) {
	// No cleanup needed for image creation step
}
