// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config

import (
	"time"

	"go.mws.cloud/go-sdk/mws"
)

const (
	DefaultSSHUsername           = "packer"
	DefaultZone                  = mws.DefaultZone
	DefaultImageDescription      = "Image created by Packer"
	DefaultDiskType              = "nbs-pl2"
	DefaultDiskSize              = "10 GB"
	DefaultDiskIOPS              = int64(1000)
	DefaultSubnetCidr            = "192.168.0.0/16"
	DefaultIPV6Prefix            = "64:ff9b::/96"
	DefaultVMType                = "gen-2-8"
	DefaultCleanupTimeout        = time.Hour
	DefaultObjectStorageRegion   = "ru-central1"
	DefaultObjectStorageEndpoint = "https://storage.mwsapis.ru"
)
