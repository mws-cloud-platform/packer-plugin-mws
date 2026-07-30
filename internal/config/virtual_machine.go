// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown

package config

import (
	"cmp"
	"errors"
	"time"
)

type VirtualMachineConfig struct {
	DiskConfig    `mapstructure:",squash"`
	NetworkConfig `mapstructure:",squash"`
	// Name for the temporary build VM (defaults to "packer-{{uuid}}-vm").
	VirtualMachineName string `mapstructure:"virtual_machine_name" required:"false"`
	// The VM type (defaults to "gen-2-8").
	VMType string `mapstructure:"vm_type" required:"false"`

	// Configuration script for initial setup of a virtual machine in the
	// [#cloud-config](https://docs.cloud-init.io/en/latest/explanation/format/cloud-config.html)
	// format. Note that this configuration would be extended with SSH key used
	// for Packer communicator.
	CloudConfig string `mapstructure:"cloud_config" required:"false"`

	// Servise account can be connected to virtual machine so that applications and scripts
	// on a virtual machine can work with MWS Cloud Platform services
	VMServiceAccount string `mapstructure:"vm_service_account" required:"false"`

	// Timeout for resources cleanup (defaults to "1h").
	CleanupTimeout time.Duration `mapstructure:"cleanup_timeout" required:"false"`
}

func (c *VirtualMachineConfig) SetDefaults() {
	c.DiskConfig.SetDefaults()
	c.NetworkConfig.SetDefaults()
	c.VMType = cmp.Or(c.VMType, DefaultVMType)
	c.CleanupTimeout = cmp.Or(c.CleanupTimeout, DefaultCleanupTimeout)
}

func (c *VirtualMachineConfig) Validate() error {
	return errors.Join(
		c.DiskConfig.Validate(),
		c.NetworkConfig.Validate(),
	)
}
