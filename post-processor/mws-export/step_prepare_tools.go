// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
)

type StepPrepareTools struct {
}

func (s *StepPrepareTools) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)
	comm := state.Get(mws.CommunicatorKey).(packer.Communicator)

	ui.Say("Checking for required tools...")

	missing, err := s.checkTools(ctx, comm, ui, "qemu-img", "aws")
	if err != nil {
		return mws.ActionHaltWithError(state, err)
	}

	if len(missing) == 0 {
		ui.Say("All required tools are already installed")
		return multistep.ActionContinue
	}

	ui.Sayf("Installing missing tools: %s", missing)

	if ok, err := s.which(ctx, comm, "apt"); err != nil {
		return mws.ActionHaltWithErrorf(state, "which apt: %w", err)
	} else if !ok {
		ui.Error("Supported package manager (apt) not found")
		return multistep.ActionHalt
	}

	if err := execTrySudo(ctx, comm, ui, "apt update"); err != nil {
		return mws.ActionHaltWithError(state, err)
	}

	for _, tool := range missing {
		if err := s.installTool(ctx, comm, ui, tool); err != nil {
			return mws.ActionHaltWithError(state, err)
		}
	}

	ui.Say("All required tools are installed")
	return multistep.ActionContinue
}

func (s *StepPrepareTools) Cleanup(multistep.StateBag) {}

func (s *StepPrepareTools) checkTools(ctx context.Context, comm packer.Communicator, ui packer.Ui, tools ...string) ([]string, error) {
	missing := make([]string, 0)
	for _, tool := range tools {
		if m, err := s.checkTool(ctx, comm, ui, tool); err != nil {
			return nil, fmt.Errorf("which %s: %w", tool, err)
		} else {
			missing = append(missing, m...)
		}
	}
	return missing, nil
}

func (s *StepPrepareTools) checkTool(ctx context.Context, comm packer.Communicator, ui packer.Ui, tool string) ([]string, error) {
	missing := make([]string, 0)

	if ok, err := s.which(ctx, comm, tool); err != nil {
		return nil, err
	} else if !ok {
		if tool == "aws" {
			// tools needed to install aws cli
			m, err := s.checkTools(ctx, comm, ui, "curl", "unzip")
			if err != nil {
				return nil, err
			}
			missing = append(missing, m...)
		}
		missing = append(missing, tool)
	}
	return missing, nil
}

func (s *StepPrepareTools) which(ctx context.Context, comm packer.Communicator, what string) (bool, error) {
	cmd := &packer.RemoteCmd{
		Command: fmt.Sprintf("which %s", what),
	}
	if err := comm.Start(ctx, cmd); err != nil {
		return false, err
	}
	return cmd.ExitStatus() == 0, nil
}

func (s *StepPrepareTools) installTool(ctx context.Context, comm packer.Communicator, ui packer.Ui, tool string) error {
	packageName := map[string]string{
		"qemu-img": "qemu-utils",
		"curl":     "curl",
		"unzip":    "unzip",
	}
	switch tool {
	case "qemu-img", "curl", "unzip":
		if err := execTrySudo(ctx, comm, ui, fmt.Sprintf("apt install -y %s", packageName[tool])); err != nil {
			return fmt.Errorf("cannot install %s", packageName[tool])
		}
	case "aws":
		if err := execTrySudo(ctx, comm, ui,
			`curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv.zip"`,
			`unzip awscliv.zip`,
			`./aws/install`,
		); err != nil {
			return fmt.Errorf("cannot install awscli")
		}
	default:
		panic("unexpected tool: " + tool)
	}
	return nil
}

func execTrySudo(ctx context.Context, comm packer.Communicator, ui packer.Ui, commands ...string) error {
	for _, command := range commands {
		outBuffer := bytes.Buffer{}
		errBuffer := bytes.Buffer{}
		cmd := &packer.RemoteCmd{
			Command: command,
			Stdout:  &outBuffer,
			Stderr:  &errBuffer,
		}

		if err := comm.Start(ctx, cmd); err != nil {
			return err
		}
		if cmd.ExitStatus() == 0 {
			for line := range strings.SplitSeq(outBuffer.String(), "\n") {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				if line != "" {
					ui.Say(line)
				}
			}
			for line := range strings.SplitSeq(errBuffer.String(), "\n") {
				line = strings.TrimRightFunc(line, unicode.IsSpace)
				if line != "" {
					ui.Error(line)
				}
			}
			continue
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
	}
	return nil
}
