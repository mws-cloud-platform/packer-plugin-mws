// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport_test

import (
	"time"

	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

const (
	prefix         = "packer-"
	project        = "test-project"
	vmName         = "packer-vm"
	cleanupTimeout = time.Hour
	errInternal    = consterr.Error("internal error")
)
