// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport

import (
	"context"
	"fmt"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
)

//nolint:revive // Very special constant for packer
const BuilderId = "packer.post-processor.mws-import"

func NewArtifact(driver Driver, image *computemodel.ImageOptionalResponse, generatedData any) *Artifact {
	return &Artifact{
		StateData: map[string]any{"generated_data": generatedData},
		driver:    driver,
		image:     image,
	}
}

type Artifact struct {
	StateData map[string]any

	driver Driver
	image  *computemodel.ImageOptionalResponse
}

//nolint:revive // Can not change packer interface
func (*Artifact) BuilderId() string {
	return BuilderId
}

func (*Artifact) Files() []string {
	return []string{}
}

//nolint:revive // Can not change packer interface
func (a *Artifact) Id() string {
	return a.image.GetMetadata().GetId().ID()
}

func (a *Artifact) String() string {
	return "Image was imported: " + a.Id()
}

func (a *Artifact) State(name string) any {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}
	data, ok := a.StateData[mws.GeneratedDataKey]
	if !ok {
		return nil
	}
	if dataMap, ok := data.(map[string]any); ok {
		return dataMap[name]
	}
	return nil
}

func (a *Artifact) Destroy() error {
	if a.driver == nil {
		return fmt.Errorf("driver is not provided in artifact: %w", mws.ErrUnexpected)
	}
	return a.driver.DeleteImage(context.Background(), string(a.image.GetMetadata().GetId().ResourceName()))
}
