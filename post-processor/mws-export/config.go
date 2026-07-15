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
	DefaultObjectStorageBucket   = "ru-central1"
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
	c.ProjectForExport = cmp.Or(c.ProjectForExport, c.Project)
}

func (c *Config) Validate() error {
	errs := append(
		c.Communicator.Prepare(&c.ctx),
		c.AccessConfig.Validate(),
		c.VirtualMachineConfig.Validate(),
		c.DiskForExportConfig.Validate(),
		c.ObjectStorageConfig.Validate(),
	)

	err := errors.Join(errs...)
	return err
}

type ObjectStorageConfig struct {
	// MWS Service Account for generation of temporal hmac-key.
	// Required, unless access_key and secret_key are provided
	ServiceAccount string `mapstructure:"service_account" required:"false"`
	// AccessKey is part of hmac-key pair for object storage.
	AccessKey string `mapstructure:"access_key" required:"false"`
	// SecretKey is part of hmac-key pair for object storage.
	SecretKey string `mapstructure:"secret_key" required:"false"`
	// Object storage region where the bucket is located (defaults to "ru-central1").
	ObjectStorageRegion string `mapstructure:"object_storage_region" required:"false"`
	// Object storage bucket where the image will be exported
	ObjectStorageBucket string `mapstructure:"object_storage_bucket" required:"true"`
	// Object storage path where the image will be stored (defaults to "packer-images/{{image_for_export}}.qcow2")
	ObjectStoragePath string `mapstructure:"object_storage_key" required:"false"`
	// Endpoint of object storage to upload image (defaults to "https://storage.mwsapis.ru").
	ObjectStorageEndpoint string `mapstructure:"object_storage_endpoint" required:"false"`
}

func (c *ObjectStorageConfig) SetDefaults() {
	c.ObjectStorageRegion = cmp.Or(c.ObjectStorageRegion, DefaultObjectStorageBucket)
	c.ObjectStorageEndpoint = cmp.Or(c.ObjectStorageEndpoint, DefaultObjectStorageEndpoint)
}

func (c *ObjectStorageConfig) Validate() error {
	if (c.SecretKey == "" || c.AccessKey == "") && c.ServiceAccount == "" {
		return consterr.Error("Object Storage authentication is not provided, " +
			"provide service_account for hmac-key generation (recommended) or pair access_key, secret_key")
	}
	return nil
}

type DiskForExportConfig struct {
	// Type of disk for export to create (defaults to "nbs-pl2").
	DiskForExportType string `mapstructure:"disk_for_export_type" required:"false"`
	// IOPS for the disk for export image (defaults to 1000).
	DiskForExportIOPS int64 `mapstructure:"disk_for_export_iops" required:"false"`
	// Project ID where the image_for_export exists (defaults to the `project`).
	ProjectForExport string `mapstructure:"project_for_export" required:"false"`
	// ID of an existing image to export (required when post processor used without mws builder).
	ImageForExport string `mapstructure:"image_for_export" required:"false"`
}

func (c *DiskForExportConfig) SetDefaults() {
	c.DiskForExportType = cmp.Or(c.DiskForExportType, commonconfig.DefaultDiskType)
	c.DiskForExportIOPS = cmp.Or(c.DiskForExportIOPS, commonconfig.DefaultDiskIOPS)
}

func (c *DiskForExportConfig) Validate() error {
	return nil
}
