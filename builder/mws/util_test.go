// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

// This file does not test anything, but contains utils for other tests.

import (
	"go.mws.cloud/go-sdk/pkg/apimodels/ipaddress"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

const (
	packerPrefix            = "packer-"
	testProjectName         = "test-project"
	testDiskName            = "test-disk"
	testExternalAddressName = "test-external-address"
	testNetworkName         = "test-network"
	testSubnetName          = "test-subnet"
	testVirtualMachineName  = "test-vm"
	testImageName           = "test-image"
	testImageDescription    = "Test image description"
	testSSHPublicKey        = "test-public-key"
	testSourceImage         = "test-source-image"

	defaultDiskName            = packerPrefix + "disk"
	defaultExternalAddressName = packerPrefix + "external-address"
	defaultNetworkName         = packerPrefix + "network"
	defaultSubnetName          = packerPrefix + "subnet"
	defaultVirtualMachineName  = packerPrefix + "vm"

	errInternal = consterr.Error("internal error")
)

var (
	testInternalAddress = new(ipaddress.MustParseIPAddressString("192.168.0.10"))
	testExternalAddress = new(ipaddress.MustParseIPAddressString("10.20.30.40"))
)
