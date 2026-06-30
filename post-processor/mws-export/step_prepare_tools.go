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

type StepPrepareTools struct {
}

func (s *StepPrepareTools) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)
	comm := state.Get(mws.CommunicatorKey).(packer.Communicator)

	var missing []string

	ui.Say("Checking for required tools...")

	for _, tool := range []string{"qemu-img", "aws"} {
		if ok, err := which(ctx, comm, tool); err != nil {
			return mws.ActionHaltWithErrorf(state, "which %s: %w", tool, err)
		} else if !ok {
			missing = append(missing, tool)
		}
	}

	if len(missing) == 0 {
		ui.Say("All required tools are already installed")
		return multistep.ActionContinue
	}

	ui.Sayf("Installing missing tools: %s", missing)

	if ok, err := which(ctx, comm, "apt"); err != nil {
		return mws.ActionHaltWithErrorf(state, "which apt: %w", err)
	} else if !ok {
		ui.Error("Supported package manager (apt) not found")
		return multistep.ActionHalt
	}

	if err := execTrySudo(ctx, comm, ui, "apt update"); err != nil {
		return mws.ActionHaltWithError(state, err)
	}

	for _, tool := range missing {
		switch tool {
		case "qemu-img":
			if err := execTrySudo(ctx, comm, ui, "apt install -y qemu-utils"); err != nil {
				return mws.ActionHaltWithError(state, err)
			}
		case "aws":
			if err := execTrySudo(ctx, comm, ui, "apt install -y awscli"); err != nil {
				return mws.ActionHaltWithError(state, err)
			}
		default:
			panic("unexpected tool: " + tool)
		}
	}

	ui.Say("All required tools are installed")
	return multistep.ActionContinue
}

func (s *StepPrepareTools) Cleanup(multistep.StateBag) {}

func which(ctx context.Context, comm packer.Communicator, what string) (bool, error) {
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("which %s", what),
	}
	if err := comm.Start(ctx, cmd); err != nil {
		return false, err
	}
	return cmd.ExitStatus() == 0, nil
}

func execTrySudo(ctx context.Context, comm packer.Communicator, ui packer.Ui, command string) error {
	cmd := &packer.RemoteCmd{
		Command: command,
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if cmd.ExitStatus() == 0 {
		return nil
	}

	ui.Sayf("Command failed with exit code %d, trying with sudo", cmd.ExitStatus())

	cmd = &packer.RemoteCmd{
		Command: "sudo " + command,
	}
	if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
		return err
	}
	if code := cmd.ExitStatus(); code != 0 {
		return fmt.Errorf("bad exit code: %d", code)
	}

	return nil
}
