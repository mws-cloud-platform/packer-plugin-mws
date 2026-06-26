// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

type PostProcessor struct {
	config Config
	runner multistep.Runner
}

func (p *PostProcessor) ConfigSpec() hcldec.ObjectSpec {
	return p.config.FlatMapstructure().HCL2Spec()
}

func (p *PostProcessor) Configure(raws ...any) error {
	err := p.config.Prepare(raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, bool, error) {
	projectForExport := ""
	imageForExport := ""
	switch artifact.BuilderId() {
	case mws.BuilderId, "packer.post-processor.artifice":
		projectForExport, _ = artifact.State("ImageProject").(string)
		imageForExport, _ = artifact.State("ImageName").(string)
	default:
		projectForExport = p.config.ProjectForExport
		imageForExport = p.config.ImageForExport
	}

	if imageForExport == "" {
		return nil, false, false, consterr.Error("image_for_export is not provided")
	}

	// prepare and render values
	var generatedData map[any]any
	stateData := artifact.State("generated_data")
	if stateData != nil {
		// Make sure it's not a nil map so we can assign to it later.
		generatedData = stateData.(map[any]any)
	}
	// If stateData has a nil map generatedData will be nil
	// and we need to make sure it's not
	if generatedData == nil {
		generatedData = make(map[any]any)
	}
	p.config.ctx.Data = generatedData

	ui.Sayf("Exporting image %s to %s", imageForExport, p.config.S3Path)

	config := mws.Config{
		PackerConfig:         p.config.PackerConfig,
		Communicator:         p.config.Communicator,
		AccessConfig:         p.config.AccessConfig,
		VirtualMachineConfig: p.config.VirtualMachineConfig,
		DiskConfig:           p.config.DiskConfig,
		NetworkConfig:        p.config.NetworkConfig,
	}

	driver, err := mws.NewDriverMWS(ctx, mws.DriverMWSConfig{
		Project:                         p.config.Project,
		BaseEndpoint:                    p.config.BaseEndpoint,
		ServiceAccountAuthorizedKeyPath: p.config.ServiceAccountAuthorizedKeyPath,
		Token:                           p.config.Token,
		CleanupTimeout:                  p.config.CleanupTimeout,
	})
	if err != nil {
		return nil, false, false, fmt.Errorf("create driver mws: %w", err)
	}

	// Prepare interpolation context
	p.config.ctx.Data = map[string]any{
		"ImageId": imageForExport,
	}

	// Interpolate S3 key if needed
	s3Path := p.config.S3Path
	if s3Path == "" {
		s3Path = fmt.Sprintf("packer-images/%s.qcow2", imageForExport)
	} else {
		ictx := p.config.ctx
		var interpolated string
		interpolated, err = interpolate.Render(s3Path, &ictx)
		if err != nil {
			return nil, false, false, fmt.Errorf("error interpolating s3_key: %w", err)
		}
		s3Path = interpolated
	}

	cloudConfig, err := mws.NewCloudConfig(config.CloudConfig)
	if err != nil {
		return nil, false, false, fmt.Errorf("create cloud config: %w", err)
	}
	cloudConfig.SetSection("package_update", true)
	// qemu-utils for saving image from disk to qcow2 file
	// unzip for unpacking aws
	cloudConfig.AppendSection("packages", "qemu-utils", "unzip")

	state := new(multistep.BasicStateBag)
	state.Put(mws.ConfigKey, &config)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UIKey, ui)
	state.Put(mws.UUIDPrefixKey, fmt.Sprintf("packer-%s-", uuid.NewString()))
	state.Put(mws.ProjectForExportKey, projectForExport)
	state.Put(mws.ImageForExportKey, imageForExport)
	state.Put(mws.S3PathKey, s3Path)
	state.Put(mws.CloudConfigKey, cloudConfig)

	steps := []multistep.Step{
		&stepPrepareS3Keys{
			S3Config: p.config.S3Config,
		},
		&communicator.StepSSHKeyGen{
			CommConf:            &config.Communicator,
			SSHTemporaryKeyPair: config.Communicator.SSHTemporaryKeyPair,
		},
		multistep.If(p.config.PackerDebug && config.Communicator.SSHPrivateKeyFile == "",
			&communicator.StepDumpSSHKey{
				Path: "mws_export" + p.config.PackerBuildName + ".pem",
				SSH:  &config.Communicator.SSH,
			},
		),
		&mws.StepCreateVirtualMachine{
			GeneratedData: &packerbuilderdata.GeneratedData{State: state},
		},
		&stepAttachDiskForExport{
			Config: p.config,
		},
		&communicator.StepConnect{
			Config:    &config.Communicator,
			Host:      mws.CommHost,
			SSHConfig: config.Communicator.SSHConfigFunc(),
		},
		&stepExportImage{
			Config: p.config,
		},
		&commonsteps.StepCleanupTempKeys{
			Comm: &config.Communicator,
		},
	}

	p.runner = commonsteps.NewRunner(steps, p.config.PackerConfig, ui)
	p.runner.Run(ctx, state)
	if rawErr, ok := state.GetOk("error"); ok {
		return nil, false, false, rawErr.(error)
	}

	url := fmt.Sprintf("https://%s/%s", DefaultS3Endpoint, strings.TrimPrefix(s3Path, "s3://"))
	result := &Artifact{
		path: s3Path,
		url:  url,
	}

	ui.Sayf("Image exported to %s", url)

	return result, false, false, nil
}
