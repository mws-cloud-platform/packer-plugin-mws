// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"cmp"
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
)

type StepCreateImage struct {
	Project          string
	ImageName        string
	ImageDisplayName string
	ImageDescription string

	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(DriverKey).(Driver)
	prefix := state.Get(PrefixKey).(string)
	ui := state.Get(UIKey).(packer.Ui)

	imageName := cmp.Or(s.ImageName, prefix+"image")

	ui.Sayf("Creating image %q from virtual machine %q...", imageName, state.Get(VirtualMachineNameKey))

	diskRef, ok := state.Get(DiskRefKey).(*computeref.DiskRef)
	if !ok || diskRef == nil {
		return ActionHaltWithErrorf(state, "disk ref not found in state: %w", ErrUnexpected)
	}

	image, err := driver.CreateImage(ctx, drivermws.CreateImageParams{
		ImageName:        imageName,
		ImageDisplayName: s.ImageDisplayName,
		ImageDescription: s.ImageDescription,
		DiskRef:          diskRef,
	})
	if err != nil {
		return ActionHaltWithErrorf(state, "create image: %w", err)
	}

	ui.Sayf("Image %q created", imageName)

	state.Put(ImageKey, image)

	s.GeneratedData.Put("ImageProject", s.Project)
	s.GeneratedData.Put("ImageName", imageName)

	return multistep.ActionContinue
}

func (*StepCreateImage) Cleanup(multistep.StateBag) {
	// No cleanup needed for image creation step
}
