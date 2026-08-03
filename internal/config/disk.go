// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown

package config

import (
	"cmp"
	"errors"
	"fmt"

	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

type DiskConfig struct {
	// Name for the disk (defaults to "packer-{{uuid}}-disk").
	DiskName string `mapstructure:"disk_name" required:"false"`
	// Type of disk to create (defaults to "nbs-pl2").
	DiskType string `mapstructure:"disk_type" required:"false"`
	// Size of the disk (defaults to "10 GB").
	DiskSize string `mapstructure:"disk_size" required:"false"`
	// IOPS for the disk (defaults to 1000).
	DiskIOPS int64 `mapstructure:"disk_iops" required:"false"`
	// Project ID where the source_image/source_disk_backup exists (defaults to the `project`).
	SourceProject string `mapstructure:"source_project" required:"false"`
	// ID of an existing image to use as a base (required unless using `source_disk_backup`).
	SourceImage string `mapstructure:"source_image" required:"false"`
	// ID of an existing disk backup to use as a base (required unless using `source_image`).
	SourceDiskBackup string `mapstructure:"source_disk_backup" required:"false"`
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
	if (c.SourceImage == "") == (c.SourceDiskBackup == "") {
		err = errors.Join(err, consterr.Error("exactly one of source_image or source_disk_backup must be provided"))
	}
	return err
}
