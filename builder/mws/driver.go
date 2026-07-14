// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"

	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -destination=mock/driver_mock.go . Driver

var _ Driver = &drivermws.Driver{}

type Driver interface {
	StepCreateVirtualMachineDriver
	CreateImage(context.Context, drivermws.CreateImageParams) (*computemodel.ImageOptionalResponse, error)
	DeleteImage(context.Context, string) error
}
