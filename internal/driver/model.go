// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
)

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
	VMType             string
	Zone               string
	SSHUsername        string
	SSHPublicKey       string
	CloudConfig        string
	DiskRef            *computeref.DiskRef
	ExternalAddressRef *vpcref.ExternalAddressRef
	SubnetRef          *vpcref.SubnetRef
}

type CreateImageParams struct {
	ImageName        string
	ImageDisplayName string
	ImageDescription string
	DiskRef          *computeref.DiskRef
}

type CreateFirewallRuleParams struct {
	NetworkName                   string
	FirewallRuleName              string
	VirtualMachineInternalAddress string
}

type ImportImageParams struct {
	ImageName        string
	ImageDisplayName string
	ImageDescription string
	ExternalURL      string
}
