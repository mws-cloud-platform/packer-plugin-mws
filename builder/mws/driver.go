// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"

	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
)

//go:generate mockgen -typed -destination=mock/driver_mock.go . Driver

type Driver interface {
	CreateDisk(context.Context, CreateDiskParams) error
	CreateExternalAddress(context.Context, CreateExternalAddressParams) (string, error)
	CreateNetwork(context.Context, CreateNetworkParams) error
	CreateSubnet(context.Context, CreateSubnetParams) error
	CreateVirtualMachine(context.Context, CreateVirtualMachineParams) (string, error)
	CreateFirewallRule(context.Context, CreateFirewallRuleParams) error
	CreateImage(context.Context, CreateImageParams) (*computemodel.ImageOptionalResponse, error)

	DeleteDisk(context.Context, string) error
	DeleteExternalAddress(context.Context, string) error
	DeleteNetwork(context.Context, string) error
	DeleteSubnet(context.Context, string, string) error
	DeleteVirtualMachine(context.Context, string) error
	DeleteFirewallRule(context.Context, string, string) error
	DeleteImage(context.Context, string) error
}

type CreateDiskParams struct {
	DiskName    string
	DiskType    string
	Size        bytesize.ByteSize
	Iops        int64
	ImageRef    *computeref.ImageRef
	SnapshotRef *computeref.SnapshotRef
	Zone        string
}

type CreateExternalAddressParams struct {
	ExternalAddressName string
}

type CreateNetworkParams struct {
	NetworkName string
}

type CreateSubnetParams struct {
	NetworkName string
	SubnetName  string
	SubnetCidr  cidraddress.CIDR4Address
}

type CreateVirtualMachineParams struct {
	VirtualMachineName string
	VmType             string
	Zone               string
	SSHUsername        string
	SSHPublicKey       string
	DiskRef            *computeref.DiskRef
	ExternalAddressRef *vpcref.ExternalAddressRef
	SubnetRef          *vpcref.SubnetRef
}

type CreateImageParams struct {
	ImageName        string
	ImageDescription string
	DiskRef          *computeref.DiskRef
}

type CreateFirewallRuleParams struct {
	NetworkName                   string
	FirewallRuleName              string
	VirtualMachineInternalAddress string
}
