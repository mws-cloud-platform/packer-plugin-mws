// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"
	"fmt"

	computemodel "go.mws.cloud/go-sdk/service/compute/model"
)

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
	data, ok := a.StateData[GeneratedDataKey]
	if !ok {
		return nil
	}
	dataMap, ok := data.(map[string]any)
	if !ok {
		return nil
	}
	return dataMap[name]
}

func (a *Artifact) Destroy() error {
	if a.driver == nil {
		return fmt.Errorf("driver is not provided in artifact: %w", errUnexpected)
	}
	return a.driver.DeleteImage(context.Background(), string(a.image.GetMetadata().GetId().ResourceName()))
}
