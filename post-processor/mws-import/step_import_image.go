// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport

import (
	"cmp"
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
)

type StepImportImage struct {
	Project          string
	ImageName        string
	ImageDisplayName string
	ImageDescription string

	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepImportImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(mws.DriverKey).(Driver)
	prefix := state.Get(mws.PrefixKey).(string)
	ui := state.Get(mws.UIKey).(packer.Ui)

	imageName := cmp.Or(s.ImageName, prefix+"image")

	externalURL, ok := state.Get(ExternalURLKey).(string)
	if !ok || externalURL == "" {
		return mws.ActionHaltWithErrorf(state, "object storage url not found in state: %w", mws.ErrUnexpected)
	}

	ui.Sayf("Importing image %q from %q...", imageName, externalURL)

	image, err := driver.ImportImage(ctx, drivermws.ImportImageParams{
		ImageName:        imageName,
		ImageDisplayName: s.ImageDisplayName,
		ImageDescription: s.ImageDescription,
		ExternalURL:      externalURL,
	})
	if err != nil {
		return mws.ActionHaltWithErrorf(state, "create image: %w", err)
	}

	ui.Sayf("Image %q imported", imageName)

	state.Put(mws.ImageKey, image)

	s.GeneratedData.Put("ImageProject", s.Project)
	s.GeneratedData.Put("ImageName", imageName)

	return multistep.ActionContinue
}

func (*StepImportImage) Cleanup(multistep.StateBag) {}
