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
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

const (
	//nolint:revive // Very special constant for packer
	BuilderId = "packer.mws"

	errUnexpected = consterr.Error("plugin unexpected error")
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
		"ImageName",
	}
	return generatedDataKeys, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	driver, err := NewDriverMWS(ctx, driverMWSConfig{
		project:                         b.config.Project,
		baseEndpoint:                    b.config.BaseEndpoint,
		serviceAccountAuthorizedKeyPath: b.config.ServiceAccountAuthorizedKeyPath,
		token:                           b.config.Token,
		cleanupTimeout:                  b.config.CleanupTimeout,
	})
	if err != nil {
		return nil, fmt.Errorf("create driver mws: %w", err)
	}

	state := new(multistep.BasicStateBag)
	state.Put(ConfigKey, &b.config)
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
			GeneratedData: generatedData,
		},
		&communicator.StepConnect{
			Config:    &b.config.Communicator,
			Host:      CommHost,
			SSHConfig: b.config.Communicator.SSHConfigFunc(),
		},
		&commonsteps.StepProvision{},
		&commonsteps.StepCleanupTempKeys{
			Comm: &b.config.Communicator,
		},
		&StepCreateImage{
			GeneratedData: generatedData,
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
		return nil, fmt.Errorf("image not found in state: %w", errUnexpected)
	}
	image, ok := v.(*computemodel.ImageOptionalResponse)
	if !ok {
		return nil, fmt.Errorf("image found in state has wrong type %T: %w", v, errUnexpected)
	}

	return &Artifact{
		StateData: map[string]any{"generated_data": state.Get(GeneratedDataKey)},
		driver:    driver,
		image:     image,
	}, nil
}
