// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"cmp"
	"context"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
)

type StepShutdownVirtualMachine struct {
	VirtualMachineName string
}

func (s *StepShutdownVirtualMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(common.DriverKey).(Driver)
	prefix := state.Get(common.PrefixKey).(string)
	ui := state.Get(common.UIKey).(packer.Ui)

	virtualMachineName := cmp.Or(s.VirtualMachineName, prefix+"vm")

	ui.Sayf("Shut down virtual machine %q...", virtualMachineName)
	err := driver.ShutdownVirtualMachine(ctx, virtualMachineName)
	if err != nil {
		return common.ActionHaltWithErrorf(state, "shut down virtual machine %q: %w", virtualMachineName, err)
	}
	ui.Sayf("Virtual machine %q had been shut down", virtualMachineName)

	return multistep.ActionContinue
}

func (s *StepShutdownVirtualMachine) Cleanup(state multistep.StateBag) {}
