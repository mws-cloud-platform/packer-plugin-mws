// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"context"

	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"go.mws.cloud/go-sdk/pkg/apimodels/ipaddress"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -destination=mock/step_create_virtual_machine_driver_mock.go . StepCreateVirtualMachineDriver

type StepCreateVirtualMachineDriver interface {
	CreateDisk(context.Context, drivermws.CreateDiskParams) error
	CreateExternalAddress(context.Context, drivermws.CreateExternalAddressParams) (*ipaddress.IPAddress, error)
	CreateNetwork(context.Context, drivermws.CreateNetworkParams) error
	CreateSubnet(context.Context, drivermws.CreateSubnetParams) error
	CreateVirtualMachine(context.Context, drivermws.CreateVirtualMachineParams) (*ipaddress.IPAddress, error)
	CreateFirewallRule(context.Context, drivermws.CreateFirewallRuleParams) error

	DeleteDisk(context.Context, string) error
	DeleteExternalAddress(context.Context, string) error
	DeleteNetwork(context.Context, string) error
	DeleteSubnet(context.Context, string, string) error
	DeleteVirtualMachine(context.Context, string) error
	DeleteFirewallRule(context.Context, string, string) error
}
