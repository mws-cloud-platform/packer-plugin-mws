// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown

package config

import (
	"cmp"
)

type ImageConfig struct {
	// Name for the resulting image (defaults to "packer-{{uuid}}-image").
	ImageName string `mapstructure:"image_name" required:"false"`
	// Display name for the resulting image (defaults to the `image_name`).
	ImageDisplayName string `mapstructure:"image_display_name" required:"false"`
	// Description for the resulting image. (defaults to "Image created by Packer").
	ImageDescription string `mapstructure:"image_description" required:"false"`
}

func (c *ImageConfig) SetDefaults() {
	c.ImageDescription = cmp.Or(c.ImageDescription, DefaultImageDescription)
}

func (c *ImageConfig) Validate() error {
	return nil
}
