// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
)

const BuilderId = "packer.mws"

type Builder struct {
	config Config
	runner multistep.Runner
}

func (b *Builder) ConfigSpec() hcldec.ObjectSpec { return b.config.FlatMapstructure().HCL2Spec() }

func (b *Builder) Prepare(raws ...interface{}) (generatedVars []string, warnings []string, err error) {
	err = config.Decode(&b.config, &config.DecodeOpts{
		PluginType:  BuilderId,
		Interpolate: true,
	}, raws...)
	if err != nil {
		return nil, nil, err
	}
	return []string{}, nil, nil
}

func (b *Builder) Run(ctx context.Context, ui packer.Ui, hook packer.Hook) (packer.Artifact, error) {
	// Setup the state bag and initial state for the steps
	state := new(multistep.BasicStateBag)
	state.Put("hook", hook)
	state.Put("ui", ui)

	// Setup builder steps.
	steps := []multistep.Step{
		// - generate SSH key pair (communicator.StepSSHKeyGen)
		// - if PackerDebug dump keys (communicator.StepDumpSSHKey)
		// - create vm
		// - connect to vm via SSH (communicator.StepConnect)
		// - run provisioners (commonsteps.StepProvision)
		// - cleanup SSH key pair (communicator.StepCleanupTempKeys)
		// - teardown vm
		// - create image
	}

	// Run!
	b.runner = commonsteps.NewRunner(steps, b.config.PackerConfig, ui)
	b.runner.Run(ctx, state)

	// If there was an error, return that
	if err, ok := state.GetOk("error"); ok {
		return nil, err.(error)
	}

	return &Artifact{
		StateData: map[string]interface{}{"generated_data": state.Get("generated_data")},
	}, nil
}
