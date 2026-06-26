// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
)

type stepExportImage struct {
	Config
}

func (s *stepExportImage) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get(mws.UIKey).(packer.Ui)
	accessKey := state.Get(mws.S3AccessKeyKey).(string)
	secretKey := state.Get(mws.S3SecretKeyKey).(string)
	s3Path := state.Get(mws.S3PathKey).(string)
	comm := state.Get("communicator").(packer.Communicator)

	if err := firstError([]func() error{
		// install aws for image upload to s3
		execCmd(ctx, comm, ui, `curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv.zip"`),
		execCmd(ctx, comm, ui, `unzip awscliv.zip`),
		execCmd(ctx, comm, ui, `sudo ./aws/install`),
		execCmd(ctx, comm, ui, `aws --version`),

		// convert raw image from disk /dev/disk/by-id/mws-image-for-export to image.qcow2 file with compression
		execCmd(ctx, comm, ui, `sudo qemu-img convert -f raw -O qcow2 -c /dev/disk/by-id/mws-image-for-export image.qcow2`),

		// TODO install certs and remove --no-verify-ssl from upload cmd
		// sudo curl -o /usr/local/share/ca-certificates/root.crt http://pki.mts.ru/root.crt
		// sudo curl -o /usr/local/share/ca-certificates/class2root.crt https://pki.mts.ru/class2root.crt
		// sudo curl -o /usr/local/share/ca-certificates/WinCAG2.crt http://pki.mts.ru/WinCAG2.crt
		// sudo curl -o /usr/local/share/ca-certificates/class2rootG2.crt https://pki.mts.ru/class2rootG2.crt
		// sudo curl -o /usr/local/share/ca-certificates/MTSWinCAG3.crt https://pki.mts.ru/MTSWinCAG3.crt
		// sudo curl -o /usr/local/share/ca-certificates/mws-root-ca.crt http://pki.mws.ru/certs/mws-root-ca.crt
		// sudo update-ca-certificates

		execCmd(ctx, comm, ui, `mkdir .aws`),

		uploadFile(comm, "~/.aws/credentials", fmt.Sprintf(
			"[default]\naws_access_key_id = %s\naws_secret_access_key = %s\n",
			accessKey, secretKey)),
		uploadFile(comm, "~/.aws/config", fmt.Sprintf(
			"[default]\nregion = %s\nendpoint_url = %s\n",
			s.S3Region, s.S3Endpoint)),

		// upload image.qcow2 to configured s3
		execCmd(ctx, comm, ui,
			fmt.Sprintf(`aws s3 cp image.qcow2 s3://%s/%s  --no-verify-ssl`, s.S3Bucket, s3Path)),
	}); err != nil {
		return mws.ActionHaltWithError(state, err)
	}

	return multistep.ActionContinue
}

func (s *stepExportImage) Cleanup(state multistep.StateBag) {}

func uploadFile(comm packer.Communicator, fileName, fileContent string) func() error {
	return func() error {
		if err := comm.Upload(fileName, strings.NewReader(fileContent), nil); err != nil {
			return fmt.Errorf("upload file %s: %w", fileName, err)
		}
		return nil
	}
}

func execCmd(ctx context.Context, comm packer.Communicator, ui packer.Ui, cmdStr string) func() error {
	return func() error {
		cmd := &packer.RemoteCmd{
			Command: cmdStr,
		}
		if err := cmd.RunWithUi(ctx, comm, ui); err != nil {
			return fmt.Errorf("executing remote command: %w", err)
		}
		if cmd.ExitStatus() != 0 {
			return fmt.Errorf("bad exit code: %d", cmd.ExitStatus())
		}
		return nil
	}
}

func firstError(funcs []func() error) error {
	for _, f := range funcs {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}
