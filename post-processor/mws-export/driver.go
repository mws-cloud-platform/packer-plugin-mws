// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"

	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/steps"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -package=mock -destination=mock/driver_mock.go . Driver

var _ Driver = &drivermws.Driver{}

type Driver interface {
	steps.StepCreateVirtualMachineDriver
	steps.StepCreateHMACKeyDriver
	CreateDisk(context.Context, drivermws.CreateDiskParams) error
	AttachDiskToVirtualMachine(ctx context.Context, vmName string, diskRef computeref.DiskRef) error
	DetachSecondaryDisksFromVirtualMachine(context.Context, string) error
	DeleteDisk(context.Context, string) error

	GetImage(context.Context, computeref.ImageRef) (*computemodel.ImageOptionalResponse, error)
}
