// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown
//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 mapstructure-to-hcl2 -type Config

package mwsexport

import (
	"cmp"
	"errors"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	commonconfig "github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

type Config struct {
	common.PackerConfig               `mapstructure:",squash"`
	Communicator                      communicator.Config `mapstructure:",squash" json:"-"`
	commonconfig.AccessConfig         `mapstructure:",squash"`
	commonconfig.VirtualMachineConfig `mapstructure:",squash"`
	commonconfig.DiskForExportConfig  `mapstructure:",squash"`
	commonconfig.ObjectStorageConfig  `mapstructure:",squash"`
	// MWS Cloud Platform Object Storage path where the image will be stored.
	ObjectStoragePath string `mapstructure:"object_storage_path" required:"true"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...any) error {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"object_storage_path",
			},
		},
	}, raws...)
	if err != nil {
		return err
	}

	c.SetDefaults()
	return c.Validate()
}

func (c *Config) SetDefaults() {
	c.Communicator.SSHUsername = cmp.Or(c.Communicator.SSHUsername, commonconfig.DefaultSSHUsername)

	c.AccessConfig.SetDefaults()
	c.VirtualMachineConfig.SetDefaults()
	c.DiskForExportConfig.SetDefaults()
	c.ObjectStorageConfig.SetDefaults()

	c.SourceProject = cmp.Or(c.SourceProject, c.Project)
	c.ImageForExportProject = cmp.Or(c.ImageForExportProject, c.Project)
}

func (c *Config) Validate() error {
	errs := append(
		c.Communicator.Prepare(&c.ctx),
		c.AccessConfig.Validate(),
		c.VirtualMachineConfig.Validate(),
		c.DiskForExportConfig.Validate(),
		c.ObjectStorageConfig.Validate(),
		interpolate.Validate(c.ObjectStoragePath, &c.ctx),
	)
	if c.ObjectStoragePath == "" {
		errs = append(errs, consterr.Error("object_storage_path is not provided"))
	}

	return errors.Join(errs...)
}
