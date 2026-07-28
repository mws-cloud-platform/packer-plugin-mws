// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown

package config

import "cmp"

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
	c.DiskForExportType = cmp.Or(c.DiskForExportType, DefaultDiskType)
	c.DiskForExportIOPS = cmp.Or(c.DiskForExportIOPS, DefaultDiskIOPS)
}

func (c *DiskForExportConfig) Validate() error {
	return nil
}
