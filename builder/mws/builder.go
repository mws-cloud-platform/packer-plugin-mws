// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

const (
	//nolint:revive // Very special constant for packer
	BuilderId = "packer.mws"

	ErrUnexpected = consterr.Error("plugin unexpected error")
)

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...any) ([]string, []string, error) {
	if err := b.config.Prepare(raws...); err != nil {
		return nil, nil, err
	}
	generatedDataKeys := []string{
		"SourceProject",
		"SourceImageName",
		"SourceSnapshotName",
		"ImageProject",
		"ImageName",
	}
	return generatedDataKeys, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	driver, err := drivermws.NewDriver(ctx, drivermws.Config{
		Project:                         b.config.Project,
		BaseEndpoint:                    b.config.BaseEndpoint,
		ServiceAccountAuthorizedKeyPath: b.config.ServiceAccountAuthorizedKeyPath,
		Token:                           b.config.Token,
		CleanupTimeout:                  b.config.CleanupTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("create driver mws: %w", err)
	}

	state := new(multistep.BasicStateBag)
	state.Put(DriverKey, driver)
	state.Put(HookKey, hook)
	state.Put(UIKey, ui)
	state.Put(PrefixKey, fmt.Sprintf("packer-%s-", uuid.NewString()))
	generatedData := &packerbuilderdata.GeneratedData{State: state}

	steps := []multistep.Step{
		&communicator.StepSSHKeyGen{
			CommConf:            &b.config.Communicator,
			SSHTemporaryKeyPair: b.config.Communicator.SSHTemporaryKeyPair,
		},
		multistep.If(b.config.PackerDebug && b.config.Communicator.SSHPrivateKeyFile == "",
			&communicator.StepDumpSSHKey{
				Path: "mws_" + b.config.PackerBuildName + ".pem",
				SSH:  &b.config.Communicator.SSH,
			},
		),
		&StepCreateVirtualMachine{
			Communicator:         &b.config.Communicator,
			AccessConfig:         b.config.AccessConfig,
			VirtualMachineConfig: b.config.VirtualMachineConfig,
			GeneratedData:        generatedData,
		},
		&communicator.StepConnect{
			Config:    &b.config.Communicator,
			Host:      CommHost(b.config.Communicator.SSHHost),
			SSHConfig: b.config.Communicator.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Communicator,
		},
		&StepCreateImage{
			Project:          b.config.Project,
			ImageName:        b.config.ImageName,
			ImageDisplayName: b.config.ImageDisplayName,
			ImageDescription: b.config.ImageDescription,
			GeneratedData:    generatedData,
		},
	}

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	if err, ok := state.GetOk(ErrorKey); ok {
		return nil, err.(error)
	}

	v, ok := state.GetOk(ImageKey)
	if !ok {
		return nil, fmt.Errorf("image not found in state: %w", ErrUnexpected)
	}
	image, ok := v.(*computemodel.ImageOptionalResponse)
	if !ok {
		return nil, fmt.Errorf("image found in state has wrong type %T: %w", v, ErrUnexpected)
	}

	result := NewArtifact(driver, image, state.Get(GeneratedDataKey))

	ui.Say(result.String())

	return result, nil
}
