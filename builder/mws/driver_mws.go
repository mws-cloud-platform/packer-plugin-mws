// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"
	"fmt"
	"time"

	"go.mws.cloud/go-sdk/mws"
	"go.mws.cloud/go-sdk/mws/wait"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	commonmodel "go.mws.cloud/go-sdk/service/common/model"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

const sshScript = `#cloud-config
users:
  - name: %s
    groups: sudo
    shell: /bin/bash
    sudo: 'ALL=(ALL) NOPASSWD:ALL'
    ssh-authorized-keys:
      - %s`

type driverMWSConfig struct {
	project                         string
	baseEndpoint                    string
	serviceAccountAuthorizedKeyPath string
	token                           string
	cleanupTimeout                  string
}

type driverMWS struct {
	disks             *computesdk.Disk
	externalAddresses *vpcsdk.ExternalAddress
	networks          *vpcsdk.Network
	subnets           *vpcsdk.Subnet
	virtualMachines   *computesdk.VirtualMachine
	firewallRules     *vpcsdk.FirewallRule
	images            *computesdk.Image
	cleanupTimeout    time.Duration
}

func NewDriverMWS(ctx context.Context, c driverMWSConfig) (Driver, error) {
	config, err := mws.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load sdk config: %w", err)
	}
	config.Project = c.project
	if c.baseEndpoint != "" {
		config.BaseEndpoint = c.baseEndpoint
	}
	if c.serviceAccountAuthorizedKeyPath != "" {
		config.ServiceAccountAuthorizedKeyPath = c.serviceAccountAuthorizedKeyPath
	}
	if c.token != "" {
		config.Token = c.token
	}

	sdk, err := mws.Load(ctx, mws.WithConfig(*config))
	if err != nil {
		return nil, fmt.Errorf("load sdk: %w", err)
	}

	disks, err := computesdk.NewDisk(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create disk client: %w", err)
	}

	externalAddresses, err := vpcsdk.NewExternalAddress(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create external address client: %w", err)
	}

	networks, err := vpcsdk.NewNetwork(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create network client: %w", err)
	}

	subnets, err := vpcsdk.NewSubnet(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create subnet client: %w", err)
	}

	virtualMachines, err := computesdk.NewVirtualMachine(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create virtual machine client: %w", err)
	}

	firewallRules, err := vpcsdk.NewFirewallRule(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create firewall rule client: %w", err)
	}

	images, err := computesdk.NewImage(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create image client: %w", err)
	}

	cleanupTimeout, err := time.ParseDuration(c.cleanupTimeout)
	if err != nil {
		return nil, fmt.Errorf("parse cleanup timeout: %w", err)
	}

	return &driverMWS{
		disks:             disks,
		externalAddresses: externalAddresses,
		networks:          networks,
		subnets:           subnets,
		virtualMachines:   virtualMachines,
		firewallRules:     firewallRules,
		images:            images,
		cleanupTimeout:    cleanupTimeout,
	}, nil
}

func (d *driverMWS) CreateDisk(ctx context.Context, params CreateDiskParams) error {
	if _, err := d.disks.CreateDisk(ctx, computeclient.UpsertDiskRequest{
		Disk: params.DiskName,
		Body: computemodel.DiskRequest{
			Metadata: &commonmodel.CommonTypedResourceMetadataRequest{
				Description: new("Disk created by Packer"),
			},
			Spec: computemodel.DiskSpecRequest{
				Zone: params.Zone,
				Size: new(params.Size),
				Source: &computemodel.DiskSpecSourceRequest{
					Image:    params.ImageRef,
					Snapshot: params.SnapshotRef,
				},
				DiskType: new(computeref.NewDiskTypeRef(params.DiskType)),
				Iops:     new(computemodel.Iops(params.Iops)),
			},
		},
	}, computeclient.WithWait()); err != nil {
		return fmt.Errorf("create disk: %w", err)
	}

	return nil
}

func (d *driverMWS) CreateExternalAddress(ctx context.Context, params CreateExternalAddressParams) (string, error) {
	resp, err := d.externalAddresses.CreateExternalAddress(ctx, vpcclient.UpsertExternalAddressRequest{
		ExternalAddress: params.ExternalAddressName,
		Body: &vpcmodel.ExternalAddressRequest{
			Metadata: &commonmodel.CommonTypedResourceMetadataRequest{
				Description: new("External address created by Packer"),
			},
			Spec: vpcmodel.VpcExternalAddressSpecRequest{},
		},
	}, vpcclient.WithWait())
	if err != nil {
		return "", fmt.Errorf("create external address: %w", err)
	}

	ipAddress := resp.Status.GetIpAddress()
	if ipAddress == nil {
		return "", consterr.Error("ip address is not available")
	}

	return ipAddress.String(), nil
}

func (d *driverMWS) CreateNetwork(ctx context.Context, params CreateNetworkParams) error {
	if _, err := d.networks.CreateNetwork(ctx, vpcclient.UpsertNetworkRequest{
		Network: params.NetworkName,
		Body: vpcmodel.NetworkRequest{
			Metadata: &commonmodel.CommonTypedResourceMetadataRequest{
				Description: new("Network created by Packer"),
			},
			Spec: vpcmodel.VpcNetworkSpecRequest{
				InternetAccess: new(true),
			},
		},
	}, vpcclient.WithWait()); err != nil {
		return fmt.Errorf("create network: %w", err)
	}

	return nil
}

func (d *driverMWS) CreateSubnet(ctx context.Context, params CreateSubnetParams) error {
	if _, err := d.subnets.CreateSubnet(ctx, vpcclient.UpsertSubnetRequest{
		Network: params.NetworkName,
		Subnet:  params.SubnetName,
		Body: vpcmodel.SubnetRequest{
			Metadata: &commonmodel.CommonTypedResourceMetadataRequest{
				Description: new("Subnet created by Packer"),
			},
			Spec: vpcmodel.SubnetSpecRequest{
				Cidr: params.SubnetCidr,
			},
		},
	}, vpcclient.WithWait()); err != nil {
		return fmt.Errorf("create subnet: %w", err)
	}

	return nil
}

func (d *driverMWS) CreateVirtualMachine(ctx context.Context, params CreateVirtualMachineParams) (string, error) {
	userData := fmt.Sprintf(sshScript, params.SSHUsername, params.SSHPublicKey)

	req := computeclient.UpsertVirtualMachineRequest{
		VirtualMachine: params.VirtualMachineName,
		Body: computemodel.VirtualMachineRequest{
			Metadata: &computemodel.VirtualMachineMetadataRequest{
				TypedResourceMetadataRequest: commonmodel.TypedResourceMetadataRequest{
					Description: new("Virtual machine created by Packer"),
				},
			},
			Spec: computemodel.VirtualMachineSpecRequest{
				Zone:   params.Zone,
				VmType: computeref.NewVmTypeRef(params.VmType),
				Hardware: &computemodel.HardwareSpecRequest{
					Power: new(computemodel.HardwareSpecPowerRequest_ON),
				},
				Os: &computemodel.OsSpecRequest{
					Metadata: &computemodel.OsSpecMetadataRequest{
						Attributes: map[string]string{
							"user-data": userData,
						},
					},
				},
				Storage: computemodel.StorageSpecRequest{
					Disks: []computemodel.StorageDiskSpecOrRefWithAttachmentsRequest{
						{
							Name: "boot",
							Boot: new(true),
							Disk: computemodel.StorageDiskSpecOrRefRequest{
								Ref: params.DiskRef,
							},
						},
					},
				},
				Network: computemodel.NetworkSpecRequest{
					NetworkInterfaces: []computemodel.NetworkInterfaceSpecRequest{
						{
							Name:    "network-interface-primary",
							Primary: new(true),
							Addresses: []computemodel.AddressSpecOrRefWithAttachmentsRequest{
								{
									Address: computemodel.AddressSpecOrRefRequest{
										Spec: &computemodel.AddressSpecRequest{
											Subnet: *params.SubnetRef,
										},
									},
									OneToOneNat: &computemodel.ComputeOneToOneNatSpecRequest{
										External: computemodel.ComputeOneToOneNatSpecExternalRequest{
											Address: computemodel.OneToOneNatAddressSpecOrRefRequest{
												Ref: params.ExternalAddressRef,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resp, err := d.virtualMachines.CreateVirtualMachine(ctx, req, computeclient.WithWait())
	if err != nil {
		return "", fmt.Errorf("create virtual machine: %w", err)
	}

	internalAddress := resp.Status.Network.GetNetworkInterfaces()[0].Addresses[0].IpAddress.String()
	return internalAddress, nil
}

func (d *driverMWS) CreateFirewallRule(ctx context.Context, params CreateFirewallRuleParams) error {
	destAddress, err := cidraddress.ParseCIDR4AddressString(params.VirtualMachineInternalAddress + "/32")
	if err != nil {
		return fmt.Errorf("parse destination CIDR for firewall rule: %w", err)
	}

	if _, err := d.firewallRules.CreateFirewallRule(ctx, vpcclient.UpsertFirewallRuleRequest{
		Network:      params.NetworkName,
		FirewallRule: params.FirewallRuleName,
		Body: vpcmodel.FirewallRuleRequest{
			Metadata: &commonmodel.CommonTypedResourceMetadataRequest{
				Description: new("Firewall rule created by Packer"),
			},
			Spec: vpcmodel.FirewallRuleSpecRequest{
				Direction: vpcmodel.FirewallRuleSpecDirectionRequest_INGRESS,
				Priority:  new(int32(1000)),
				Action:    vpcmodel.FirewallRuleSpecActionRequest_ALLOW,
				Source: vpcmodel.FirewallRuleSourceRequest{
					Spec: &vpcmodel.FirewallRuleSourceSpecRequest{
						Cidrs: []cidraddress.CIDR4Address{
							cidraddress.MustParseCIDR4AddressString("0.0.0.0/0"),
						},
					},
				},
				Destination: vpcmodel.FirewallRuleDestinationRequest{
					Spec: &vpcmodel.FirewallRuleDestinationSpecRequest{
						Cidrs: []cidraddress.CIDR4Address{
							destAddress,
						},
					},
				},
			},
		},
	}, vpcclient.WithWait()); err != nil {
		return fmt.Errorf("create firewall rule: %w", err)
	}

	return nil
}

func (d *driverMWS) CreateImage(ctx context.Context, params CreateImageParams) (*computemodel.ImageOptionalResponse, error) {
	image, err := d.images.CreateImage(ctx, computeclient.UpsertImageRequest{
		Image: params.ImageName,
		Body: computemodel.ImageRequest{
			Metadata: &commonmodel.CommonTypedResourceMetadataRequest{
				Description: new(params.ImageDescription),
			},
			Spec: computemodel.ImageSpecRequest{
				Source: computemodel.ImageSpecSourceRequest{
					DiskId: params.DiskRef,
				},
			},
		},
	}, computeclient.WithWait())
	if err != nil {
		return nil, fmt.Errorf("create image: %w", err)
	}

	return image, nil
}

func (d *driverMWS) DeleteDisk(ctx context.Context, diskName string) error {
	if err := d.disks.DeleteDisk(ctx, computeclient.DeleteDiskRequest{
		Disk: diskName,
	}, computeclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete disk: %w", err)
	}

	return nil
}

func (d *driverMWS) DeleteExternalAddress(ctx context.Context, externalAddressName string) error {
	if err := d.externalAddresses.DeleteExternalAddress(ctx, vpcclient.DeleteExternalAddressRequest{
		ExternalAddress: externalAddressName,
	}, vpcclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete external address: %w", err)
	}

	return nil
}

func (d *driverMWS) DeleteNetwork(ctx context.Context, networkName string) error {
	if err := d.networks.DeleteNetwork(ctx, vpcclient.DeleteNetworkRequest{
		Network: networkName,
	}, vpcclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete network: %w", err)
	}

	return nil
}

func (d *driverMWS) DeleteSubnet(ctx context.Context, networkName, subnetName string) error {
	if err := d.subnets.DeleteSubnet(ctx, vpcclient.DeleteSubnetRequest{
		Network: networkName,
		Subnet:  subnetName,
	}, vpcclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete subnet: %w", err)
	}

	return nil
}

func (d *driverMWS) DeleteVirtualMachine(ctx context.Context, virtualMachineName string) error {
	if err := d.virtualMachines.DeleteVirtualMachine(ctx, computeclient.DeleteVirtualMachineRequest{
		VirtualMachine: virtualMachineName,
	}, computeclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete virtual machine: %w", err)
	}

	return nil
}

func (d *driverMWS) DeleteFirewallRule(ctx context.Context, networkName, firewallRuleName string) error {
	if err := d.firewallRules.DeleteFirewallRule(ctx, vpcclient.DeleteFirewallRuleRequest{
		Network:      networkName,
		FirewallRule: firewallRuleName,
	}, vpcclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete firewall rule: %w", err)
	}

	return nil
}

func (d *driverMWS) DeleteImage(ctx context.Context, imageName string) error {
	if err := d.images.DeleteImage(ctx, computeclient.DeleteImageRequest{
		Image: imageName,
	}, computeclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	return nil
}
