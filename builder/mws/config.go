// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown
//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 mapstructure-to-hcl2 -type Config

package mws

import (
	"cmp"
	"errors"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
)

type Config struct {
	common.PackerConfig  `mapstructure:",squash"`
	Communicator         communicator.Config `mapstructure:",squash" json:"-"`
	AccessConfig         `mapstructure:",squash"`
	ImageConfig          `mapstructure:",squash"`
	VirtualMachineConfig `mapstructure:",squash"`
	DiskConfig           `mapstructure:",squash"`
	NetworkConfig        `mapstructure:",squash"`

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
	c.ImageConfig.SetDefaults()
	c.VirtualMachineConfig.SetDefaults()
	c.DiskConfig.SetDefaults()
	c.NetworkConfig.SetDefaults()

	c.SourceProject = cmp.Or(c.SourceProject, c.Project)
}

func (c *Config) Validate() error {
	errs := append(
		c.Communicator.Prepare(&c.ctx),
		c.AccessConfig.Validate(),
		c.ImageConfig.Validate(),
		c.VirtualMachineConfig.Validate(),
		c.DiskConfig.Validate(),
		c.NetworkConfig.Validate(),
	)

	err := errors.Join(errs...)
	return err
}
