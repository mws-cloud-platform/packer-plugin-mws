// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package mws

import "github.com/hashicorp/packer-plugin-sdk/common"

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
}
