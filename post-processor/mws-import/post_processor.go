// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/multistep/commonsteps"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	mwsexport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
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

func (p *PostProcessor) PostProcess(ctx context.Context, ui packer.Ui, _ packer.Artifact) (packer.Artifact, bool, bool, error) {
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

	state := new(multistep.BasicStateBag)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UIKey, ui)
	state.Put(mws.PrefixKey, fmt.Sprintf("packer-%s-", uuid.NewString()))
	generatedData := &packerbuilderdata.GeneratedData{State: state}

	steps := []multistep.Step{
		&mwsexport.StepCreateHMACKey{
			AccessKey:      p.config.AccessKey,
			SecretKey:      p.config.SecretKey,
			ServiceAccount: p.config.ServiceAccount,
			CleanupTimeout: p.config.CleanupTimeout,
		},
		&StepCreateSignedLink{
			Endpoint: p.config.ObjectStorageEndpoint,
			Region:   p.config.ObjectStorageRegion,
			Path:     p.config.ObjectStoragePath,
		},
		&StepImportImage{
			Project:          p.config.Project,
			ImageName:        p.config.ImageName,
			ImageDisplayName: p.config.ImageDisplayName,
			ImageDescription: p.config.ImageDescription,
			GeneratedData:    generatedData,
		},
	}

	p.runner = commonsteps.NewRunner(steps, p.config.PackerConfig, ui)
	p.runner.Run(ctx, state)
	if rawErr, ok := state.GetOk(mws.ErrorKey); ok {
		return nil, false, false, rawErr.(error)
	}

	v, ok := state.GetOk(mws.ImageKey)
	if !ok {
		return nil, false, false, fmt.Errorf("image not found in state: %w", mws.ErrUnexpected)
	}
	image, ok := v.(*computemodel.ImageOptionalResponse)
	if !ok {
		return nil, false, false, fmt.Errorf("image found in state has wrong type %T: %w", v, mws.ErrUnexpected)
	}

	result := NewArtifact(driver, image, state.Get(mws.GeneratedDataKey))

	ui.Say(result.String())

	return result, false, false, nil
}
