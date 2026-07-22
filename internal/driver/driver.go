// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"context"
	"fmt"
	"time"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/cloudinit"
	"go.mws.cloud/go-sdk/mws"
	"go.mws.cloud/go-sdk/mws/wait"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/ipaddress"
	"go.mws.cloud/go-sdk/pkg/optional"
	commonmodel "go.mws.cloud/go-sdk/service/common/model"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"
	iamclient "go.mws.cloud/go-sdk/service/iam/client"
	iammodel "go.mws.cloud/go-sdk/service/iam/model"
	iamsdk "go.mws.cloud/go-sdk/service/iam/sdk"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcclient "go.mws.cloud/go-sdk/service/vpc/client"
	vpcmodel "go.mws.cloud/go-sdk/service/vpc/model"
	vpcsdk "go.mws.cloud/go-sdk/service/vpc/sdk"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

type Driver struct {
	disks             *computesdk.Disk
	externalAddresses *vpcsdk.ExternalAddress
	networks          *vpcsdk.Network
	subnets           *vpcsdk.Subnet
	virtualMachines   *computesdk.VirtualMachine
	firewallRules     *vpcsdk.FirewallRule
	images            *computesdk.Image
	hmacKeys          *iamsdk.ServiceAccountHmacKey

	cleanupTimeout time.Duration
}

func NewDriver(ctx context.Context, c Config) (*Driver, error) {
	config, err := mws.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("load sdk config: %w", err)
	}

	config.Project = c.Project
	if c.BaseEndpoint != "" {
		config.BaseEndpoint = c.BaseEndpoint
	}
	if c.ServiceAccountAuthorizedKeyPath != "" {
		config.ServiceAccountAuthorizedKeyPath = c.ServiceAccountAuthorizedKeyPath
	}
	if c.Token != "" {
		config.Token = c.Token
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

	hmacKeys, err := iamsdk.NewServiceAccountHmacKey(ctx, sdk)
	if err != nil {
		return nil, fmt.Errorf("create hmac keys client: %w", err)
	}

	return &Driver{
		disks:             disks,
		externalAddresses: externalAddresses,
		networks:          networks,
		subnets:           subnets,
		virtualMachines:   virtualMachines,
		firewallRules:     firewallRules,
		images:            images,
		hmacKeys:          hmacKeys,
		cleanupTimeout:    c.CleanupTimeout,
	}, nil
}

func (d *Driver) CreateDisk(ctx context.Context, params CreateDiskParams) error {
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

func (d *Driver) CreateExternalAddress(ctx context.Context, params CreateExternalAddressParams) (*ipaddress.IPAddress, error) {
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
		return nil, fmt.Errorf("create external address: %w", err)
	}

	ipAddress := resp.Status.GetIpAddress()
	if ipAddress == nil {
		return nil, consterr.Error("ip address is not available")
	}

	return ipAddress, nil
}

func (d *Driver) CreateNetwork(ctx context.Context, params CreateNetworkParams) error {
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

func (d *Driver) CreateSubnet(ctx context.Context, params CreateSubnetParams) error {
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

func (d *Driver) CreateVirtualMachine(ctx context.Context, params CreateVirtualMachineParams) (*ipaddress.IPAddress, error) {
	userData, err := cloudinit.PrepareCloudConfig(params.SSHUsername, params.SSHPublicKey, params.CloudConfig)
	if err != nil {
		return nil, fmt.Errorf("prepare cloud-config: %w", err)
	}

	var oneToOneNat *computemodel.ComputeOneToOneNatSpecRequest
	if params.ExternalAddressRef != nil {
		oneToOneNat = &computemodel.ComputeOneToOneNatSpecRequest{
			External: computemodel.ComputeOneToOneNatSpecExternalRequest{
				Address: computemodel.OneToOneNatAddressSpecOrRefRequest{
					Ref: params.ExternalAddressRef,
				},
			},
		}
	}

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
				VmType: computeref.NewVmTypeRef(params.VMType),
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
									OneToOneNat: oneToOneNat,
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
		return nil, fmt.Errorf("create virtual machine: %w", err)
	}

	internalAddress := resp.Status.Network.GetNetworkInterfaces()[0].Addresses[0].IpAddress
	return internalAddress, nil
}

func (d *Driver) CreateFirewallRule(ctx context.Context, params CreateFirewallRuleParams) error {
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
				ProtoPorts: []string{"TCP:22"},
			},
		},
	}, vpcclient.WithWait()); err != nil {
		return fmt.Errorf("create firewall rule: %w", err)
	}

	return nil
}

func (d *Driver) CreateImage(ctx context.Context, params CreateImageParams) (*computemodel.ImageOptionalResponse, error) {
	var displayName *string
	if params.ImageDisplayName != "" {
		displayName = new(params.ImageDisplayName)
	}
	image, err := d.images.CreateImage(ctx, computeclient.UpsertImageRequest{
		Image: params.ImageName,
		Body: computemodel.ImageRequest{
			Metadata: &commonmodel.CommonTypedResourceMetadataRequest{
				Description: new(params.ImageDescription),
				DisplayName: displayName,
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

func (d *Driver) CreateHMACKey(ctx context.Context, serviceAccount, name string) (string, string, error) {
	hmacKey, err := d.hmacKeys.CreateHmacKey(ctx, iamclient.UpsertHmacKeyRequest{
		ServiceAccount: serviceAccount,
		KeyName:        name,
		Body: iammodel.HmacKeyRequest{
			Spec: iammodel.HmacKeySpecRequest{
				ExpirationTime: new(time.Now().Add(time.Hour)),
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("create hmac key: %w", err)
	}

	return hmacKey.GetStatus().GetAccessKeyIdOr(""), hmacKey.GetStatus().GetSecretAccessKeyOr(""), nil
}

func (d *Driver) GetImage(ctx context.Context, imageRef computeref.ImageRef) (*computemodel.ImageOptionalResponse, error) {
	image, err := d.images.GetImage(ctx, computeclient.GetImageRequest{
		Project: imageRef.GetProject(),
		Image:   imageRef.GetImage(),
	}, computeclient.WithWait())
	if err != nil {
		return nil, fmt.Errorf("get image: %w", err)
	}

	return image, nil
}

func (d *Driver) AttachDiskToVirtualMachine(ctx context.Context, vmName string, diskRef computeref.DiskRef) error {
	_, err := d.virtualMachines.UpdateVirtualMachine(ctx, computeclient.UpdateVirtualMachineRequest{
		VirtualMachine: vmName,
		Body: computemodel.UpdateVirtualMachineRequest{
			Spec: optional.NewOptional(computemodel.UpdateVirtualMachineSpecRequest{
				Storage: optional.NewOptional(computemodel.UpdateStorageSpecRequest{
					Disks: optional.NewOptional([]computemodel.UpdateStorageDiskSpecOrRefWithAttachmentsRequest{
						{Name: optional.NewOptional("boot")},
						{
							Name: optional.NewOptional(DiskForExportName),
							Disk: optional.NewOptional(computemodel.UpdateStorageDiskSpecOrRefRequest{
								Ref: optional.NewOptional(diskRef),
							}),
						},
					}),
				}),
			}),
		},
	}, computeclient.WithWait())
	if err != nil {
		return fmt.Errorf("attach disk to virtual machine: %w", err)
	}
	return nil
}

func (d *Driver) DetachSecondaryDisksFromVirtualMachine(ctx context.Context, virtualMachineName string) error {
	_, err := d.virtualMachines.UpdateVirtualMachine(ctx, computeclient.UpdateVirtualMachineRequest{
		VirtualMachine: virtualMachineName,
		Body: computemodel.UpdateVirtualMachineRequest{
			Spec: optional.NewOptional(computemodel.UpdateVirtualMachineSpecRequest{
				Storage: optional.NewOptional(computemodel.UpdateStorageSpecRequest{
					Disks: optional.NewOptional([]computemodel.UpdateStorageDiskSpecOrRefWithAttachmentsRequest{
						{Name: optional.NewOptional("boot")},
					}),
				}),
			}),
		},
	}, computeclient.WithWait())
	if err != nil {
		return fmt.Errorf("detach disk from virtual machine: %w", err)
	}
	return nil
}

func (d *Driver) DeleteDisk(ctx context.Context, diskName string) error {
	if err := d.disks.DeleteDisk(ctx, computeclient.DeleteDiskRequest{
		Disk: diskName,
	}, computeclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete disk: %w", err)
	}

	return nil
}

func (d *Driver) DeleteExternalAddress(ctx context.Context, externalAddressName string) error {
	if err := d.externalAddresses.DeleteExternalAddress(ctx, vpcclient.DeleteExternalAddressRequest{
		ExternalAddress: externalAddressName,
	}, vpcclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete external address: %w", err)
	}

	return nil
}

func (d *Driver) DeleteNetwork(ctx context.Context, networkName string) error {
	if err := d.networks.DeleteNetwork(ctx, vpcclient.DeleteNetworkRequest{
		Network: networkName,
	}, vpcclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete network: %w", err)
	}

	return nil
}

func (d *Driver) DeleteSubnet(ctx context.Context, networkName, subnetName string) error {
	if err := d.subnets.DeleteSubnet(ctx, vpcclient.DeleteSubnetRequest{
		Network: networkName,
		Subnet:  subnetName,
	}, vpcclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete subnet: %w", err)
	}

	return nil
}

func (d *Driver) DeleteVirtualMachine(ctx context.Context, virtualMachineName string) error {
	if err := d.virtualMachines.DeleteVirtualMachine(ctx, computeclient.DeleteVirtualMachineRequest{
		VirtualMachine: virtualMachineName,
	}, computeclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete virtual machine: %w", err)
	}

	return nil
}

func (d *Driver) DeleteFirewallRule(ctx context.Context, networkName, firewallRuleName string) error {
	if err := d.firewallRules.DeleteFirewallRule(ctx, vpcclient.DeleteFirewallRuleRequest{
		Network:      networkName,
		FirewallRule: firewallRuleName,
	}, vpcclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete firewall rule: %w", err)
	}

	return nil
}

func (d *Driver) DeleteImage(ctx context.Context, imageName string) error {
	if err := d.images.DeleteImage(ctx, computeclient.DeleteImageRequest{
		Image: imageName,
	}, computeclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete image: %w", err)
	}

	return nil
}

func (d *Driver) DeleteHMACKey(ctx context.Context, serviceAccount string, name string) error {
	if err := d.hmacKeys.DeleteHmacKey(ctx, iamclient.DeleteHmacKeyRequest{
		ServiceAccount: serviceAccount,
		KeyName:        name,
	}, iamclient.WithWait(wait.WithTimeout(d.cleanupTimeout))); err != nil {
		return fmt.Errorf("delete hmac key: %w", err)
	}

	return nil
}
