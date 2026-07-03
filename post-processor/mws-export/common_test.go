// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport_test

import (
	"errors"
	"time"
)

const (
	prefix         = "packer-"
	project        = "test-project"
	vmName         = "packer-vm"
	cleanupTimeout = time.Hour
)

var errInternal = errors.New("internal error")
