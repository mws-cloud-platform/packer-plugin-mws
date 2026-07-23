// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown
//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 mapstructure-to-hcl2 -type Config

package mwsimport

import (
	"cmp"
	"errors"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/common"
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
	common.PackerConfig       `mapstructure:",squash"`
	commonconfig.AccessConfig `mapstructure:",squash"`
	commonconfig.ImageConfig  `mapstructure:",squash"`
	ObjectStorageConfig       `mapstructure:",squash"`

	// Timeout for resources cleanup (defaults to "1h").
	CleanupTimeout time.Duration `mapstructure:"cleanup_timeout" required:"false"`

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
	c.AccessConfig.SetDefaults()
	c.ImageConfig.SetDefaults()
	c.ObjectStorageConfig.SetDefaults()
	c.CleanupTimeout = cmp.Or(c.CleanupTimeout, commonconfig.DefaultCleanupTimeout)
}

func (c *Config) Validate() error {
	err := errors.Join(
		c.AccessConfig.Validate(),
		c.ImageConfig.Validate(),
		c.ObjectStorageConfig.Validate(),
	)
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

	// MWS Cloud Platform Object Storage path from where the image will be imported.
	ObjectStoragePath string `mapstructure:"object_storage_path" required:"true"`
	// MWS Cloud Platform Object Storage endpoint to import image from (defaults to "https://storage.mwsapis.ru").
	ObjectStorageEndpoint string `mapstructure:"object_storage_endpoint" required:"false"`
	// MWS Cloud Platform Object Storage region where the bucket is located (defaults to "ru-central1").
	ObjectStorageRegion string `mapstructure:"object_storage_region" required:"false"`
}

func (c *ObjectStorageConfig) SetDefaults() {
	c.ObjectStorageEndpoint = cmp.Or(c.ObjectStorageEndpoint, DefaultObjectStorageEndpoint)
	c.ObjectStorageRegion = cmp.Or(c.ObjectStorageRegion, DefaultObjectStorageRegion)
}

func (c *ObjectStorageConfig) Validate() error {
	if c.ObjectStoragePath == "" {
		return consterr.Error("object_storage_path is not provided")
	}
	if (c.SecretKey == "" || c.AccessKey == "") && c.ServiceAccount == "" {
		return consterr.Error("Object Storage authentication is not provided, " +
			"provide service_account for hmac-key generation (recommended) or pair access_key, secret_key")
	}
	return nil
}
