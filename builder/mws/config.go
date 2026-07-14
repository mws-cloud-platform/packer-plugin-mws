// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown
//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 mapstructure-to-hcl2 -type Config

package mws

import (
	"cmp"
	"errors"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	commonconfig "github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
	"go.mws.cloud/go-sdk/mws"
)

const (
	DefaultSSHUsername      = "packer"
	DefaultZone             = mws.DefaultZone
	DefaultVMType           = "gen-2-8"
	DefaultDiskType         = "nbs-pl2"
	DefaultDiskSize         = "10 GB"
	DefaultDiskIOPS         = int64(1000)
	DefaultSubnetCidr       = "192.168.0.0/16"
	DefaultImageDescription = "Image created by Packer"
	DefaultCleanupTimeout   = time.Hour
)

type Config struct {
	common.PackerConfig               `mapstructure:",squash"`
	Communicator                      communicator.Config `mapstructure:",squash" json:"-"`
	commonconfig.AccessConfig         `mapstructure:",squash"`
	commonconfig.DiskConfig           `mapstructure:",squash"`
	commonconfig.ImageConfig          `mapstructure:",squash"`
	commonconfig.NetworkConfig        `mapstructure:",squash"`
	commonconfig.VirtualMachineConfig `mapstructure:",squash"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...any) error {
	if err := config.Decode(c, &config.DecodeOpts{
		PluginType:         BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
	}, raws...); err != nil {
		return err
	}

	c.SetDefaults()
	return c.Validate()
}

func (c *Config) SetDefaults() {
	c.Communicator.SSHUsername = cmp.Or(c.Communicator.SSHUsername, DefaultSSHUsername)

	c.AccessConfig.SetDefaults()
	c.DiskConfig.SetDefaults()
	c.ImageConfig.SetDefaults()
	c.NetworkConfig.SetDefaults()
	c.VirtualMachineConfig.SetDefaults()

	c.SourceProject = cmp.Or(c.SourceProject, c.Project)
}

func (c *Config) Validate() error {
	errs := append(
		c.Communicator.Prepare(&c.ctx),
		c.AccessConfig.Validate(),
		c.DiskConfig.Validate(),
		c.ImageConfig.Validate(),
		c.NetworkConfig.Validate(),
		c.VirtualMachineConfig.Validate(),
	)

	err := errors.Join(errs...)
	return err
}
