//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown
//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 mapstructure-to-hcl2 -type Config

package mwsexport

import (
	"cmp"
	"errors"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

const (
	DefaultS3Bucket   = "ru-central1"
	DefaultS3Endpoint = "https://storage.mwsapis.ru"
)

type Config struct {
	common.PackerConfig      `mapstructure:",squash"`
	Communicator             communicator.Config `mapstructure:",squash" json:"-"`
	mws.AccessConfig         `mapstructure:",squash"`
	mws.VirtualMachineConfig `mapstructure:",squash"`
	mws.DiskConfig           `mapstructure:",squash"`
	mws.NetworkConfig        `mapstructure:",squash"`

	DiskForExportConfig `mapstructure:",squash"`
	S3Config            `mapstructure:",squash"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) error {
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
	c.Communicator.SSHUsername = cmp.Or(c.Communicator.SSHUsername, mws.DefaultSSHUsername)

	c.AccessConfig.SetDefaults()
	c.VirtualMachineConfig.SetDefaults()
	c.DiskConfig.SetDefaults()
	c.NetworkConfig.SetDefaults()
	c.DiskForExportConfig.SetDefaults()
	c.S3Config.SetDefaults()

	c.SourceProject = cmp.Or(c.SourceProject, c.Project)
	c.ProjectForExport = cmp.Or(c.ProjectForExport, c.Project)
}

func (c *Config) Validate() error {
	errs := append(
		c.Communicator.Prepare(&c.ctx),
		c.AccessConfig.Validate(),
		c.VirtualMachineConfig.Validate(),
		c.DiskConfig.Validate(),
		c.NetworkConfig.Validate(),
		c.DiskForExportConfig.Validate(),
		c.S3Config.Validate(),
	)

	err := errors.Join(errs...)
	return err
}

type S3Config struct {
	// MWS Service Account for generation of temporal hmac-key.
	// Required, unless access_key and secret_key are provided
	ServiceAccount string `mapstructure:"service_account" required:"false"`
	// AccessKey is part of hmac-key pair for S3.
	AccessKey string `mapstructure:"access_key" required:"false"`
	// SecretKey is part of hmac-key pair for S3.
	SecretKey string `mapstructure:"secret_key" required:"false"`
	// S3 region where the bucket is located (defaults to "ru-central1").
	S3Region string `mapstructure:"s3_region" required:"false"`
	// S3 bucket where the image will be exported
	S3Bucket string `mapstructure:"s3_bucket" required:"true"`
	// S3 path where the image will be stored (defaults to "packer-images/{{image_for_export}}.qcow2")
	S3Path string `mapstructure:"s3_key" required:"false"`
	// Endpoint of S3 to upload image (defaults to "https://storage.mwsapis.ru").
	S3Endpoint string `mapstructure:"s3_endpoint" required:"false"`
}

func (c *S3Config) SetDefaults() {
	c.S3Region = cmp.Or(c.S3Region, DefaultS3Bucket)
	// c.S3Path = cmp.Or(c.S3Path, mws.DefaultDiskIOPS)
	c.S3Endpoint = cmp.Or(c.S3Endpoint, DefaultS3Endpoint)
}

func (c *S3Config) Validate() error {
	if (c.SecretKey == "" || c.AccessKey == "") && c.ServiceAccount == "" {
		return consterr.Error("S3 authentication is not provided, provide service_account for hmac-key generation (recommended) or pair access_key, secret_key")
	}
	return nil
}

type DiskForExportConfig struct {
	// Type of disk for export to create (defaults to "nbs-pl2").
	DiskForExportType string `mapstructure:"disk_for_export_type" required:"false"`
	// Additional size of the disk for export image (defaults to "0 GB").
	// Adds to image_for_export minDiskSize.
	AdditionalDiskForExportSize string `mapstructure:"additional_disk_for_export_size" required:"false"`
	// IOPS for the disk for export image (defaults to 1000).
	DiskForExportIOPS int64 `mapstructure:"disk_for_export_iops" required:"false"`
	// Project ID where the image_for_export exists (defaults to the `project`).
	ProjectForExport string `mapstructure:"project_for_export" required:"false"`
	// ID of an existing image to export (required when post processor used without mws builder).
	ImageForExport string `mapstructure:"image_for_export" required:"false"`
}

func (c *DiskForExportConfig) SetDefaults() {
	c.DiskForExportType = cmp.Or(c.DiskForExportType, mws.DefaultDiskType)
	c.DiskForExportIOPS = cmp.Or(c.DiskForExportIOPS, mws.DefaultDiskIOPS)
	c.AdditionalDiskForExportSize = cmp.Or(c.AdditionalDiskForExportSize, mws.DefaultAdditionalDiskForExportSize)
}

func (c *DiskForExportConfig) Validate() error {
	var err error
	if _, parseErr := bytesize.ParseString(c.AdditionalDiskForExportSize); parseErr != nil {
		err = errors.Join(err, fmt.Errorf("parse additional_disk_for_export_size: %w", parseErr))
	}
	return err
}
