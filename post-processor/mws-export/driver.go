// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/steps"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -package=mock -destination=mock/driver_mock.go . Driver

var _ Driver = &drivermws.Driver{}

type Driver interface {
	steps.StepCreateVirtualMachineDriver
	steps.StepCreateHMACKeyDriver
}
