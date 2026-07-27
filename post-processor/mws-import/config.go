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
	commonconfig.ObjectStorageConfig `mapstructure:",squash"`
	// MWS Cloud Platform Object Storage path from where the image will be imported.
	ObjectStoragePath string `mapstructure:"object_storage_path" required:"true"`
}

func (c *ObjectStorageConfig) SetDefaults() {
	c.ObjectStorageConfig.SetDefaults()
}

func (c *ObjectStorageConfig) Validate() error {
	err := c.ObjectStorageConfig.Validate()
	if c.ObjectStoragePath == "" {
		err = errors.Join(err, consterr.Error("object_storage_path is not provided"))
	}
	return err
}
