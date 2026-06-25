// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"bytes"
	"context"
	"fmt"

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

	// install tools:
	// qemu-utils for saving image from disk to qcow2 file
	// unzip for unpacking aws
	// aws for image upload to s3
	execCmd(ctx, comm, ui, `sudo apt-get update`)
	execCmd(ctx, comm, ui, `sudo apt-get install -y qemu-utils unzip`)
	execCmd(ctx, comm, ui, `curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv.zip"`)
	execCmd(ctx, comm, ui, `unzip awscliv.zip`)
	execCmd(ctx, comm, ui, `sudo ./aws/install`)

	// convert raw image from disk /dev/disk/by-id/mws-image-for-export to image.qcow2 file with compression
	execCmd(ctx, comm, ui, `sudo qemu-img convert -f raw -O qcow2 -c /dev/disk/by-id/mws-image-for-export image.qcow2`)

	// TODO install certs and remove --no-verify-ssl from upload cmd
	// sudo curl -o /usr/local/share/ca-certificates/root.crt http://pki.mts.ru/root.crt
	// sudo curl -o /usr/local/share/ca-certificates/class2root.crt https://pki.mts.ru/class2root.crt
	// sudo curl -o /usr/local/share/ca-certificates/WinCAG2.crt http://pki.mts.ru/WinCAG2.crt
	// sudo curl -o /usr/local/share/ca-certificates/class2rootG2.crt https://pki.mts.ru/class2rootG2.crt
	// sudo curl -o /usr/local/share/ca-certificates/MTSWinCAG3.crt https://pki.mts.ru/MTSWinCAG3.crt
	// sudo curl -o /usr/local/share/ca-certificates/mws-root-ca.crt http://pki.mws.ru/certs/mws-root-ca.crt
	// sudo update-ca-certificates

	// upload image.qcow2 to configured s3
	execCmd(ctx, comm, ui, fmt.Sprintf(`AWS_ACCESS_KEY_ID="%s" AWS_SECRET_ACCESS_KEY="%s" AWS_DEFAULT_REGION="%s" aws s3 cp image.qcow2 s3://%s/%s --endpoint-url %s --no-verify-ssl`, accessKey, secretKey, s.S3Region, s.S3Bucket, s3Path, s.S3Endpoint))

	return multistep.ActionContinue
}

func (s *stepExportImage) Cleanup(state multistep.StateBag) {}

func execCmd(ctx context.Context, comm packer.Communicator, ui packer.Ui, cmdStr string) error {
	buf := bytes.Buffer{}
	cmd := &packer.RemoteCmd{
		Command: cmdStr,
		Stdout:  &buf,
		Stderr:  &buf,
	}
	if err := comm.Start(ctx, cmd); err != nil {
		return fmt.Errorf("executing remote command: %w", err)
	}
	badCode := cmd.Wait() != 0
	ui.Say(buf.String())
	if badCode {
		return fmt.Errorf("bad exit code: %d", cmd.ExitStatus())
	}
	return nil
}
