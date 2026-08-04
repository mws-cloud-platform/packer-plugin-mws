// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"cmp"
	"context"
	"math/big"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	commonconfig "github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/ipaddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	iamref "go.mws.cloud/go-sdk/service/resources/references/iam"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
)

type StepCreateVirtualMachine struct {
	Project      string
	Zone         string
	Communicator *communicator.Config
	commonconfig.VirtualMachineConfig
	*commonconfig.DiskForExportConfig

	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateVirtualMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(common.DriverKey).(StepCreateVirtualMachineDriver)
	prefix := state.Get(common.PrefixKey).(string)
	ui := state.Get(common.UIKey).(packer.Ui)

	exportDiskName := prefix + "disk-for-export"
	bootDiskName := cmp.Or(s.DiskName, prefix+"boot-disk")
	externalAddressName := cmp.Or(s.ExternalAddressName, prefix+"external-address")
	networkName := cmp.Or(s.NetworkName, prefix+"network")
	subnetName := cmp.Or(s.SubnetName, prefix+"subnet")
	virtualMachineName := cmp.Or(s.VirtualMachineName, prefix+"vm")
	firewallRuleName := prefix + "ssh-access"

	externalAddress := &ipaddress.IPAddress{}
	internalAddress := &ipaddress.IPAddress{}

	bootDiskSize, err := s.bootDiskSize(ctx, driver)
	if err != nil {
		return common.ActionHaltWithErrorf(state, "calculate boot disk size: %w", err)
	}

	if err = (subSteps{
		&subStepWithoutResult[drivermws.CreateDiskParams]{
			cond:         s.DiskForExportConfig != nil,
			resourceType: "export disk",
			resourceName: exportDiskName,
			action:       driver.CreateDisk,
			params:       s.exportDiskParams(exportDiskName),
		},
		&subStepWithoutResult[drivermws.CreateDiskParams]{
			cond:         true,
			resourceType: "boot disk",
			resourceName: bootDiskName,
			action:       driver.CreateDisk,
			params: drivermws.CreateDiskParams{
				DiskName:      bootDiskName,
				DiskType:      s.DiskType,
				Size:          bootDiskSize,
				Iops:          s.DiskIOPS,
				ImageRef:      s.bootImageRef(),
				DiskBackupRef: s.bootDiskBackupRef(),
				Zone:          s.Zone,
			},
		},
		&subStepWithResult[ipaddress.IPAddress, drivermws.CreateExternalAddressParams]{
			cond:         s.UseExternalAddress,
			resourceType: "external address",
			resourceName: externalAddressName,
			action:       driver.CreateExternalAddress,
			result:       externalAddress,
			params: drivermws.CreateExternalAddressParams{
				ExternalAddressName: externalAddressName,
			},
		},
		&subStepWithoutResult[drivermws.CreateNetworkParams]{
			cond:         s.NetworkName == "",
			resourceType: "network",
			resourceName: networkName,
			action:       driver.CreateNetwork,
			params: drivermws.CreateNetworkParams{
				NetworkName: networkName,
			},
		},
		&subStepWithoutResult[drivermws.CreateSubnetParams]{
			cond:         s.SubnetName == "",
			resourceType: "subnet",
			resourceName: subnetName,
			action:       driver.CreateSubnet,
			params: drivermws.CreateSubnetParams{
				NetworkName: networkName,
				SubnetName:  subnetName,
				SubnetCidr:  cidraddress.MustParseCIDR4AddressString(s.SubnetCidr),
			},
		},
		&subStepWithResult[ipaddress.IPAddress, drivermws.CreateVirtualMachineParams]{
			cond:         true,
			resourceType: "virtual machine",
			resourceName: virtualMachineName,
			action:       driver.CreateVirtualMachine,
			result:       internalAddress,
			params: drivermws.CreateVirtualMachineParams{
				VirtualMachineName: virtualMachineName,
				VMType:             s.VMType,
				Zone:               s.Zone,
				SSHUsername:        s.Communicator.SSHUsername,
				SSHPublicKey:       string(s.Communicator.SSHPublicKey),
				CloudConfig:        s.CloudConfig,
				ServiceAccountRef:  s.serviceAccountRef(),
				BootDiskRef:        new(computeref.NewDiskRef(s.Project, bootDiskName)),
				ExportDiskRef:      s.exportDiskRef(exportDiskName),
				ExternalAddressRef: s.externalAddressRef(externalAddressName),
				SubnetRef:          new(vpcref.NewSubnetRef(s.Project, networkName, subnetName)),
			},
		},
		&subStepWithoutResult[drivermws.CreateFirewallRuleParams]{
			cond:         s.UseExternalAddress,
			resourceType: "firewall rule",
			resourceName: firewallRuleName,
			action:       driver.CreateFirewallRule,
			params: drivermws.CreateFirewallRuleParams{
				NetworkName:                   networkName,
				FirewallRuleName:              firewallRuleName,
				VirtualMachineInternalAddress: internalAddress,
			},
		},
	}.run(ctx, ui)); err != nil {
		return common.ActionHaltWithError(state, err)
	}

	var virtualMachineAddress *ipaddress.IPAddress
	if s.UseExternalAddress {
		virtualMachineAddress = externalAddress
	} else {
		virtualMachineAddress = internalAddress
	}

	if s.Nat64Enable {
		virtualMachineAddress, err = ConvertToIPv6(virtualMachineAddress, s.Nat64IPV6Prefix)
		if err != nil {
			return common.ActionHaltWithErrorf(state, "convert virtual machine ip to IPv6: %w", err)
		}
	}
	state.Put(common.InstanceIPKey, virtualMachineAddress.String())

	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put(common.InstanceIDKey, virtualMachineName)

	// for create image step
	state.Put(common.DiskRefKey, new(computeref.NewDiskRef(s.Project, bootDiskName)))

	s.GeneratedData.Put("SourceProject", s.SourceProject)
	s.GeneratedData.Put("SourceImageName", s.SourceImage)
	s.GeneratedData.Put("SourceDiskBackupName", s.SourceDiskBackup)

	return multistep.ActionContinue
}

func (s *StepCreateVirtualMachine) Cleanup(state multistep.StateBag) {
	driver := state.Get(common.DriverKey).(StepCreateVirtualMachineDriver)
	prefix := state.Get(common.PrefixKey).(string)
	ui := state.Get(common.UIKey).(packer.Ui)

	ctx, cancel := context.WithTimeout(context.Background(), s.CleanupTimeout)
	defer cancel()

	networkName := cmp.Or(s.NetworkName, prefix+"network")

	if s.UseExternalAddress {
		firewallRuleName := prefix + "ssh-access"
		common.DeleteSubWithUI(ctx, ui, "firewall rule", firewallRuleName, "network", networkName, driver.DeleteFirewallRule)
	}

	virtualMachineName := cmp.Or(s.VirtualMachineName, prefix+"vm")
	common.DeleteWithUI(ctx, ui, "virtual machine", virtualMachineName, driver.DeleteVirtualMachine)

	if s.SubnetName == "" {
		subnetName := cmp.Or(s.SubnetName, prefix+"subnet")
		common.DeleteSubWithUI(ctx, ui, "subnet", subnetName, "network", networkName, driver.DeleteSubnet)
	}

	if s.NetworkName == "" {
		common.DeleteWithUI(ctx, ui, "network", networkName, driver.DeleteNetwork)
	}

	if s.UseExternalAddress {
		externalAddressName := cmp.Or(s.ExternalAddressName, prefix+"external-address")
		common.DeleteWithUI(ctx, ui, "external address", externalAddressName, driver.DeleteExternalAddress)
	}

	bootDiskName := cmp.Or(s.DiskName, prefix+"boot-disk")
	common.DeleteWithUI(ctx, ui, "boot disk", bootDiskName, driver.DeleteDisk)

	if s.DiskForExportConfig != nil {
		exportDiskName := prefix + "disk-for-export"
		common.DeleteWithUI(ctx, ui, "export disk", exportDiskName, driver.DeleteDisk)
	}
}

func (s *StepCreateVirtualMachine) bootDiskSize(ctx context.Context, driver StepCreateVirtualMachineDriver) (*bytesize.ByteSize, error) {
	if s.DiskSize != "" {
		// Validated in DiskConfig.Validate()
		return new(bytesize.MustParseString(s.DiskSize)), nil
	}

	exportSize, err := driver.GetImageMinDiskSize(ctx, s.exportImageRef())
	if err != nil {
		return nil, err
	}
	var bootSize *bytesize.ByteSize
	if s.SourceImage != "" {
		bootSize, err = driver.GetImageMinDiskSize(ctx, s.bootImageRef())
		if err != nil {
			return nil, err
		}
	} else {
		bootSize, err = driver.GetDiskBackupMinDiskSize(ctx, s.bootDiskBackupRef())
		if err != nil {
			return nil, err
		}
	}

	// add bootSize and exportSize
	sum := new(big.Int).Add(bootSize.BigInt(), exportSize.BigInt())
	// convert to GB
	divGB, modGB := new(big.Int).DivMod(sum, bytesize.GB.GetValue(), new(big.Int))
	// round up if necessary
	if modGB.Sign() > 0 {
		divGB.Add(divGB, big.NewInt(1))
	}

	result, err := bytesize.NewFromBigInt(divGB, bytesize.GB)
	return &result, err
}

func (s *StepCreateVirtualMachine) exportImageRef() *computeref.ImageRef {
	if s.DiskForExportConfig != nil && s.ImageForExport != "" {
		return new(computeref.NewImageRef(s.ImageForExportProject, s.ImageForExport))
	}
	return nil
}

func (s *StepCreateVirtualMachine) bootImageRef() *computeref.ImageRef {
	if s.SourceImage != "" {
		return new(computeref.NewImageRef(s.SourceProject, s.SourceImage))
	}
	return nil
}

func (s *StepCreateVirtualMachine) bootDiskBackupRef() *computeref.DiskBackupRef {
	if s.SourceDiskBackup != "" {
		return new(computeref.NewDiskBackupRef(s.SourceProject, s.SourceDiskBackup))
	}
	return nil
}

func (s *StepCreateVirtualMachine) exportDiskRef(exportDiskName string) *computeref.DiskRef {
	if s.DiskForExportConfig != nil {
		return new(computeref.NewDiskRef(s.Project, exportDiskName))
	}
	return nil
}

func (s *StepCreateVirtualMachine) serviceAccountRef() *iamref.ServiceAccountRef {
	if s.VMServiceAccount != "" {
		return new(iamref.NewServiceAccountRef(s.Project, s.VMServiceAccount))
	}
	return nil
}

func (s *StepCreateVirtualMachine) externalAddressRef(externalAddressName string) *vpcref.ExternalAddressRef {
	if s.UseExternalAddress {
		return new(vpcref.NewExternalAddressRef(s.Project, externalAddressName))
	}
	return nil
}

// separeted from other subStep params because of nil pointer dereference in DiskForExportConfig fields
func (s *StepCreateVirtualMachine) exportDiskParams(exportDiskName string) drivermws.CreateDiskParams {
	if s.DiskForExportConfig == nil {
		return drivermws.CreateDiskParams{}
	}
	return drivermws.CreateDiskParams{
		DiskName: exportDiskName,
		DiskType: s.DiskForExportType,
		Iops:     s.DiskForExportIOPS,
		ImageRef: s.exportImageRef(),
		Zone:     s.Zone,
	}
}
