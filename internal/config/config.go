// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"cmp"
	"errors"
	"fmt"
	"time"

	"go.mws.cloud/go-sdk/mws"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
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

	DefaultAdditionalDiskForExportSize = "0 GB"
)

type AccessConfig struct {
	// The project identifier where resources will be created.
	Project string `mapstructure:"project" required:"true"`
	// The zone in which the VM will be created (defaults to "ru-central1-a")
	Zone string `mapstructure:"zone" required:"false"`

	// MWS Cloud Platform API base endpoint (defaults to "https://api.mwsapis.ru").
	// Can be specified using the `MWS_BASE_ENDPOINT` environment variable.
	BaseEndpoint string `mapstructure:"base_endpoint" required:"false"`
	// Path to the service account authorized key file used for authentication.
	// Has no effect if IAM token is set.
	// Can be specified using the `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` environment variable.
	ServiceAccountAuthorizedKeyPath string `mapstructure:"service_account_authorized_key_path" required:"false"`
	// IAM token used for authentication.
	// Can be specified using the `MWS_TOKEN` environment variable.
	Token string `mapstructure:"token" required:"false"`
}

func (c *AccessConfig) SetDefaults() {
	c.Zone = cmp.Or(c.Zone, DefaultZone)
}

func (c *AccessConfig) Validate() error {
	if c.Project == "" {
		return consterr.Error("project is not provided")
	}
	return nil
}

type ImageConfig struct {
	// Name for the resulting image (defaults to "packer-{{uuid}}-image").
	ImageName string `mapstructure:"image_name" required:"false"`
	// Description for the resulting image. (defaults to "Image created by Packer").
	ImageDescription string `mapstructure:"image_description" required:"false"`
}

func (c *ImageConfig) SetDefaults() {
	c.ImageDescription = cmp.Or(c.ImageDescription, DefaultImageDescription)
}

func (c *ImageConfig) Validate() error {
	return nil
}

type VirtualMachineConfig struct {
	DiskConfig    `mapstructure:",squash"`
	NetworkConfig `mapstructure:",squash"`
	// Name for the temporary build VM (defaults to "packer-{{uuid}}-vm").
	VirtualMachineName string `mapstructure:"virtual_machine_name" required:"false"`
	// The VM type (defaults to "gen-2-8").
	VMType string `mapstructure:"vm_type" required:"false"`

	// Timeout for cleanup of create virtual machine step (defaults to "1h").
	CleanupTimeout time.Duration `mapstructure:"cleanup_timeout" required:"false"`

	// Configuration script for initial setup of a virtual machine in the
	// [#cloud-config](https://docs.cloud-init.io/en/latest/explanation/format/cloud-config.html)
	// format. Note that this configuration would be extended with SSH key used
	// for Packer communicator.
	CloudConfig string `mapstructure:"cloud_config" required:"false"`
}

func (c *VirtualMachineConfig) SetDefaults() {
	c.VMType = cmp.Or(c.VMType, DefaultVMType)
	c.CleanupTimeout = cmp.Or(c.CleanupTimeout, DefaultCleanupTimeout)
}

func (c *VirtualMachineConfig) Validate() error {
	return nil
}

type DiskConfig struct {
	// Name for the disk (defaults to "packer-{{uuid}}-disk").
	DiskName string `mapstructure:"disk_name" required:"false"`
	// Type of disk to create (defaults to "nbs-pl2").
	DiskType string `mapstructure:"disk_type" required:"false"`
	// Size of the disk (defaults to "10 GB").
	DiskSize string `mapstructure:"disk_size" required:"false"`
	// IOPS for the disk (defaults to 1000).
	DiskIOPS int64 `mapstructure:"disk_iops" required:"false"`
	// Project ID where the source image/snapshot exists (defaults to the `project`).
	SourceProject string `mapstructure:"source_project" required:"false"`
	// ID of an existing image to use as a base (required unless using `source_snapshot`).
	SourceImage string `mapstructure:"source_image" required:"false"`
	// ID of an existing snapshot to use as a base (required unless using `source_image`).
	SourceSnapshot string `mapstructure:"source_snapshot" required:"false"`
}

func (c *DiskConfig) SetDefaults() {
	c.DiskType = cmp.Or(c.DiskType, DefaultDiskType)
	c.DiskIOPS = cmp.Or(c.DiskIOPS, DefaultDiskIOPS)
	c.DiskSize = cmp.Or(c.DiskSize, DefaultDiskSize)
}

func (c *DiskConfig) Validate() error {
	var err error
	if _, parseErr := bytesize.ParseString(c.DiskSize); parseErr != nil {
		err = errors.Join(err, fmt.Errorf("parse disk size: %w", parseErr))
	}
	if (c.SourceImage == "") == (c.SourceSnapshot == "") {
		err = errors.Join(err, consterr.Error("exactly one of source_image or source_snapshot must be provided"))
	}
	return err
}

type NetworkConfig struct {
	// Name for the network (defaults to "packer-{{uuid}}-network").
	// If specified, Packer will use existing network.
	NetworkName string `mapstructure:"network_name" required:"false"`
	// Name for the subnet (defaults to "packer-{{uuid}}-subnet").
	// If specified, Packer will use existing subnet.
	SubnetName string `mapstructure:"subnet_name" required:"false"`
	// Subnet CIDR (defaults to "192.168.0.0/16").
	SubnetCidr string `mapstructure:"subnet_cidr" required:"false"`
	// Use external address for connection to virtual machine from internet (defaults to "false").
	UseExternalAddress bool `mapstructure:"use_external_address" required:"false"`
	// External address name (defaults to "packer-{{uuid}}-external-address").
	// Can be specified only if external address usage is enabled.
	ExternalAddressName string `mapstructure:"external_address_name" required:"false"`
}

func (c *NetworkConfig) SetDefaults() {
	c.SubnetCidr = cmp.Or(c.SubnetCidr, DefaultSubnetCidr)
}

func (c *NetworkConfig) Validate() error {
	var err error
	if _, parseErr := cidraddress.ParseCIDR4AddressString(c.SubnetCidr); parseErr != nil {
		err = errors.Join(err, fmt.Errorf("parse subnet CIDR: %w", parseErr))
	}
	if c.SubnetName != "" && c.NetworkName == "" {
		err = errors.Join(err, consterr.Error("when subnet_name is provided, network_name must be provided"))
	}
	if !c.UseExternalAddress && c.SubnetName == "" {
		err = errors.Join(err, consterr.Error("when use_external_address is false, subnet_name must be provided"))
	}
	if !c.UseExternalAddress && c.ExternalAddressName != "" {
		err = errors.Join(err, consterr.Error("when use_external_address is false, external_address_name must not be provided"))
	}
	return err
}
