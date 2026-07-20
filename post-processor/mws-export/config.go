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

const (
	DefaultObjectStorageRegion   = "ru-central1"
	DefaultObjectStorageEndpoint = "https://storage.mwsapis.ru"
)

type Config struct {
	common.PackerConfig               `mapstructure:",squash"`
	Communicator                      communicator.Config `mapstructure:",squash" json:"-"`
	commonconfig.AccessConfig         `mapstructure:",squash"`
	commonconfig.VirtualMachineConfig `mapstructure:",squash"`

	DiskForExportConfig `mapstructure:",squash"`
	ObjectStorageConfig `mapstructure:",squash"`

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

	err := errors.Join(errs...)
	return err
}

type ObjectStorageConfig struct {
	// MWS Cloud Platform Service Account used for generating temporal HMAC key
	// to access Object Storage. Required, unless `access_key` and `secret_key`
	// are provided.
	ServiceAccount string `mapstructure:"service_account" required:"false"`

	// HMAC key identifier for authenticating with Object Storage. Used if
	// `service_account` is not provided. Also requires `secret_key` to be
	// provided.
	AccessKey string `mapstructure:"access_key" required:"false"`
	// HMAC key secret for accessing Object Storage. Required if `access_key` is
	// provided.
	SecretKey string `mapstructure:"secret_key" required:"false"`

	// MWS Cloud Platform Object Storage path where the image will be stored.
	ObjectStoragePath string `mapstructure:"object_storage_path" required:"true"`
	// MWS Cloud Platform Object Storage endpoint to upload image (defaults to "https://storage.mwsapis.ru").
	ObjectStorageEndpoint string `mapstructure:"object_storage_endpoint" required:"false"`
	// MWS Cloud Platform Object Storage region where the bucket is located (defaults to "ru-central1").
	ObjectStorageRegion string `mapstructure:"object_storage_region" required:"false"`
}

func (c *ObjectStorageConfig) SetDefaults() {
	c.ObjectStorageEndpoint = cmp.Or(c.ObjectStorageEndpoint, DefaultObjectStorageEndpoint)
	c.ObjectStorageRegion = cmp.Or(c.ObjectStorageRegion, DefaultObjectStorageRegion)
}

func (c *ObjectStorageConfig) Validate() error {
	if (c.SecretKey == "" || c.AccessKey == "") && c.ServiceAccount == "" {
		return consterr.Error("Object Storage authentication is not provided, " +
			"provide service_account for hmac-key generation (recommended) or pair access_key, secret_key")
	}
	return nil
}

type DiskForExportConfig struct {
	// Type of the disk used for image export (defaults to "nbs-pl2").
	DiskForExportType string `mapstructure:"disk_for_export_type" required:"false"`
	// IOPS for the disk used for image export (defaults to 1000).
	DiskForExportIOPS int64 `mapstructure:"disk_for_export_iops" required:"false"`
	// The project identifier where the image for export exists (defaults to the `project`).
	ImageForExportProject string `mapstructure:"image_for_export_project" required:"false"`
	// Identifier of the image to export. Required only when post processor used
	// without mws builder.
	ImageForExport string `mapstructure:"image_for_export" required:"false"`
}

func (c *DiskForExportConfig) SetDefaults() {
	c.DiskForExportType = cmp.Or(c.DiskForExportType, commonconfig.DefaultDiskType)
	c.DiskForExportIOPS = cmp.Or(c.DiskForExportIOPS, commonconfig.DefaultDiskIOPS)
}

func (c *DiskForExportConfig) Validate() error {
	return nil
}
