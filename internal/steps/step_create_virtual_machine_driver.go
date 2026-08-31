// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"context"

	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"go.mws.cloud/go-sdk/pkg/apimodels/ipaddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -destination=mock/step_create_virtual_machine_driver_mock.go . StepCreateVirtualMachineDriver

var _ StepCreateVirtualMachineDriver = &drivermws.Driver{}

type StepCreateVirtualMachineDriver interface {
	CreateDisk(context.Context, drivermws.CreateDiskParams) error
	CreateExternalAddress(context.Context, drivermws.CreateExternalAddressParams) (*ipaddress.IPAddress, error)
	CreateNetwork(context.Context, drivermws.CreateNetworkParams) error
	CreateSubnet(context.Context, drivermws.CreateSubnetParams) error
	CreateVirtualMachine(context.Context, drivermws.CreateVirtualMachineParams) (*ipaddress.IPAddress, error)
	CreateFirewallRule(context.Context, drivermws.CreateFirewallRuleParams) error

	GetImageMinDiskSize(context.Context, *computeref.ImageRef) (*bytesize.ByteSize, error)
	GetDiskBackupMinDiskSize(context.Context, *computeref.DiskBackupRef) (*bytesize.ByteSize, error)
	GetSerialPortOutput(context.Context, string, int) ([]byte, error)

	AttachDiskToVirtualMachine(context.Context, string, *computeref.DiskRef) error

	DeleteDisk(context.Context, string) error
	DeleteExternalAddress(context.Context, string) error
	DeleteNetwork(context.Context, string) error
	DeleteSubnet(context.Context, string, string) error
	DeleteVirtualMachine(context.Context, string) error
	DeleteFirewallRule(context.Context, string, string) error
}
