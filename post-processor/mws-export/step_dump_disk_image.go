// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
)

const (
	device        = "/dev/disk/by-id/mws-disk-for-export"
	diskImageFile = "image.qcow2"
)

type StepDumpDiskImage struct {
}

func (s *StepDumpDiskImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)
	comm := state.Get(mws.CommunicatorKey).(packer.Communicator)

	sudo := ""
	checkAccessCmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("qemu-img info %s", device),
	}

	ui.Say("Checking access...")
	if err := checkAccessCmd.RunWithUi(ctx, comm, ui); err != nil {
		return mws.ActionHaltWithError(state, err)
	}
	if code := checkAccessCmd.ExitStatus(); code != 0 {
		ui.Sayf("Check access failed with exit code %d, trying to dump disk image with sudo", code)
		sudo = "sudo "
	}

	dumpCmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("%sqemu-img convert -f raw -O qcow2 -c %s %s", sudo, device, diskImageFile),
	}

	ui.Say("Dumping disk image...")
	if err := dumpCmd.RunWithUi(ctx, comm, ui); err != nil {
		return mws.ActionHaltWithError(state, err)
	}
	if code := dumpCmd.ExitStatus(); code != 0 {
		return mws.ActionHaltWithErrorf(state, "bad exit code: %d", code)
	}

	ui.Say("Disk image dumped")
	return multistep.ActionContinue
}

func (s *StepDumpDiskImage) Cleanup(multistep.StateBag) {}
