// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"cmp"
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
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
	driver := state.Get(common.DriverKey).(Driver)
	prefix := state.Get(common.PrefixKey).(string)
	ui := state.Get(common.UIKey).(packer.Ui)

	imageName := cmp.Or(s.ImageName, prefix+"image")

	ui.Sayf("Creating image %q from virtual machine %q...", imageName, state.Get(common.InstanceIDKey))

	diskRef, ok := state.Get(common.DiskRefKey).(*computeref.DiskRef)
	if !ok || diskRef == nil {
		return common.ActionHaltWithErrorf(state, "disk ref not found in state: %w", common.ErrUnexpected)
	}

	image, err := driver.CreateImage(ctx, drivermws.CreateImageParams{
		ImageName:        imageName,
		ImageDisplayName: s.ImageDisplayName,
		ImageDescription: s.ImageDescription,
		DiskRef:          diskRef,
	})
	if err != nil {
		return common.ActionHaltWithErrorf(state, "create image: %w", err)
	}

	ui.Sayf("Image %q created", imageName)

	state.Put(common.ImageKey, image)

	s.GeneratedData.Put("ImageProject", s.Project)
	s.GeneratedData.Put("ImageName", imageName)

	return multistep.ActionContinue
}

func (*StepCreateImage) Cleanup(multistep.StateBag) {}
