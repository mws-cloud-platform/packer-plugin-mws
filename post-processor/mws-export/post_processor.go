// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

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
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/steps"
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
	// prepare values for interpolation
	var generatedData map[any]any
	stateData := artifact.State(common.GeneratedDataKey)
	if stateData != nil {
		generatedData = stateData.(map[any]any)
	}
	if generatedData == nil {
		generatedData = make(map[any]any)
	}

	diskForExportConfig := p.config.DiskForExportConfig
	switch artifact.BuilderId() {
	case "packer.mws", "packer.post-processor.artifice":
		diskForExportConfig.ImageForExportProject, _ = artifact.State("ImageProject").(string)
		diskForExportConfig.ImageForExport, _ = artifact.State("ImageName").(string)
	default:
		// put values for interpolation in case of exporting image without builder
		generatedData["ImageProject"] = diskForExportConfig.ImageForExportProject
		generatedData["ImageName"] = diskForExportConfig.ImageForExport
	}

	if diskForExportConfig.ImageForExport == "" {
		return nil, false, false, consterr.Error("image_for_export is not provided")
	}

	p.config.ctx.Data = generatedData
	// interpolate values
	objectStoragePath, err := interpolate.Render(p.config.ObjectStoragePath, &p.config.ctx)
	if err != nil {
		return nil, false, false, fmt.Errorf("interpolate object_storage_path: %w", err)
	}

	driver, err := driver.NewDriver(ctx, driver.Config{
		Project:                         p.config.Project,
		BaseEndpoint:                    p.config.BaseEndpoint,
		ServiceAccountAuthorizedKeyPath: p.config.ServiceAccountAuthorizedKeyPath,
		Token:                           p.config.Token,
		CleanupTimeout:                  p.config.CleanupTimeout,
	})
	if err != nil {
		return nil, false, false, fmt.Errorf("create driver mws: %w", err)
	}

	ui.Sayf("Exporting image %s to %s", diskForExportConfig.ImageForExport, objectStoragePath)

	state := new(multistep.BasicStateBag)
	state.Put(common.DriverKey, driver)
	state.Put(common.UIKey, ui)
	state.Put(common.PrefixKey, fmt.Sprintf("packer-%s-", uuid.NewString()))

	steps := []multistep.Step{
		&communicator.StepSSHKeyGen{
			CommConf:            &p.config.Communicator,
			SSHTemporaryKeyPair: p.config.Communicator.SSHTemporaryKeyPair,
		},
		multistep.If(p.config.PackerDebug && p.config.Communicator.SSHPrivateKeyFile == "",
			&communicator.StepDumpSSHKey{
				Path: "mws_export" + p.config.PackerBuildName + ".pem",
				SSH:  &p.config.Communicator.SSH,
			},
		),
		&steps.StepCreateHMACKey{
			AccessKey:      p.config.AccessKey,
			SecretKey:      p.config.SecretKey,
			ServiceAccount: p.config.ServiceAccount,
			CleanupTimeout: p.config.CleanupTimeout,
		},
		&steps.StepCreateVirtualMachine{
			Project:              p.config.Project,
			Zone:                 p.config.Zone,
			Communicator:         &p.config.Communicator,
			VirtualMachineConfig: p.config.VirtualMachineConfig,
			DiskForExportConfig:  &diskForExportConfig,
			GeneratedData:        &packerbuilderdata.GeneratedData{State: state},
		},
		&communicator.StepConnect{
			Config:    &p.config.Communicator,
			Host:      common.CommHost(p.config.Communicator.SSHHost),
			SSHConfig: p.config.Communicator.SSHConfigFunc(),
		},
		&StepPrepareTools{},
		&StepDumpDiskImage{},
		&StepUploadAWSSharedCredsFile{},
		&StepUploadImage{
			Endpoint: p.config.ObjectStorageEndpoint,
			Region:   p.config.ObjectStorageRegion,
			Path:     objectStoragePath,
		},
		&commonsteps.StepCleanupTempKeys{
			Comm: &p.config.Communicator,
		},
	}

	p.runner = commonsteps.NewRunner(steps, p.config.PackerConfig, ui)
	p.runner.Run(ctx, state)
	if rawErr, ok := state.GetOk(common.ErrorKey); ok {
		return nil, false, false, rawErr.(error)
	}

	url := fmt.Sprintf("%s/%s", p.config.ObjectStorageEndpoint, objectStoragePath)
	result := &Artifact{
		path: objectStoragePath,
		url:  url,
	}

	ui.Say(result.String())

	return result, false, false, nil
}
