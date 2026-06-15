// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"cmp"
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
	config := state.Get(ConfigKey).(*Config)
	driver := state.Get(DriverKey).(Driver)
	prefix := state.Get(UuidPrefixKey).(string)
	ui := state.Get(UiKey).(packer.Ui)

	imageName := cmp.Or(config.ImageName, prefix+"image")

	ui.Sayf("Creating image %q from virtual machine %q...", imageName, state.Get(VirtualMachineNameKey))

	diskRef, ok := state.Get(DiskRefKey).(*computeref.DiskRef)
	if !ok || diskRef == nil {
		return actionHaltWithError(state, fmt.Errorf("disk ref not found in state: %w", errUnexpected))
	}

	image, err := driver.CreateImage(ctx, CreateImageParams{
		ImageName:        imageName,
		ImageDescription: config.ImageDescription,
		DiskRef:          diskRef,
	})
	if err != nil {
		return actionHaltWithError(state, fmt.Errorf("create image: %w", err))
	}

	ui.Sayf("Image %q created", imageName)

	state.Put(ImageKey, image)

	s.GeneratedData.Put("ImageName", imageName)

	return multistep.ActionContinue
}

func (*StepCreateImage) Cleanup(multistep.StateBag) {
	// No cleanup needed for image creation step
}
