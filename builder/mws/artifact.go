// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"
	"fmt"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
)

func NewArtifact(driver Driver, image *computemodel.ImageOptionalResponse, generatedData any) *Artifact {
	return &Artifact{
		StateData: map[string]any{common.GeneratedDataKey: generatedData},
		driver:    driver,
		image:     image,
	}
}

type Artifact struct {
	// StateData should store data such as GeneratedData
	// to be shared with post-processors
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
	return "A disk image was created: " + a.Id()
}

func (a *Artifact) State(name string) any {
	if _, ok := a.StateData[name]; ok {
		return a.StateData[name]
	}
	data, ok := a.StateData[common.GeneratedDataKey]
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
		return fmt.Errorf("driver is not provided in artifact: %w", common.ErrUnexpected)
	}
	return a.driver.DeleteImage(context.Background(), string(a.image.GetMetadata().GetId().ResourceName()))
}
