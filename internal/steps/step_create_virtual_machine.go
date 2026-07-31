// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"cmp"
	"context"

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

	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateVirtualMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(common.DriverKey).(StepCreateVirtualMachineDriver)
	prefix := state.Get(common.PrefixKey).(string)
	ui := state.Get(common.UIKey).(packer.Ui)

	var (
		externalAddressRef    *vpcref.ExternalAddressRef
		virtualMachineAddress *ipaddress.IPAddress
	)

	diskName := cmp.Or(s.DiskName, prefix+"disk")
	ui.Sayf("Creating disk...")
	if err := driver.CreateDisk(ctx, drivermws.CreateDiskParams{
		DiskName:    diskName,
		DiskType:    s.DiskType,
		Size:        bytesize.MustParseString(s.DiskSize),
		Iops:        s.DiskIOPS,
		ImageRef:    s.imageRef(),
		SnapshotRef: s.snapshotRef(),
		Zone:        s.Zone,
	}); err != nil {
		return common.ActionHaltWithErrorf(state, "create disk %q: %w", diskName, err)
	}
	ui.Sayf("Disk %q created", diskName)

	diskRef := new(computeref.NewDiskRef(s.Project, diskName))
	state.Put(common.DiskRefKey, diskRef)

	if s.UseExternalAddress {
		externalAddressName := cmp.Or(s.ExternalAddressName, prefix+"external-address")
		ui.Sayf("Creating external address...")
		externalAddress, err := driver.CreateExternalAddress(ctx, drivermws.CreateExternalAddressParams{
			ExternalAddressName: externalAddressName,
		})
		if err != nil {
			return common.ActionHaltWithErrorf(state, "create external-address %q: %w", externalAddressName, err)
		}

		ui.Sayf("External Address %q created", externalAddressName)
		virtualMachineAddress = externalAddress
		externalAddressRef = new(vpcref.NewExternalAddressRef(s.Project, externalAddressName))
	}

	networkName := cmp.Or(s.NetworkName, prefix+"network")
	if s.NetworkName == "" {
		ui.Sayf("Creating network...")
		if err := driver.CreateNetwork(ctx, drivermws.CreateNetworkParams{
			NetworkName: networkName,
		}); err != nil {
			return common.ActionHaltWithErrorf(state, "create network %q: %w", networkName, err)
		}

		ui.Sayf("Network %q created", networkName)
	}

	subnetName := cmp.Or(s.SubnetName, prefix+"subnet")
	if s.SubnetName == "" {
		ui.Sayf("Creating subnet...")
		if err := driver.CreateSubnet(ctx, drivermws.CreateSubnetParams{
			NetworkName: networkName,
			SubnetName:  subnetName,
			SubnetCidr:  cidraddress.MustParseCIDR4AddressString(s.SubnetCidr),
		}); err != nil {
			return common.ActionHaltWithErrorf(state, "create subnet %q: %w", subnetName, err)
		}

		ui.Sayf("Subnet %q created", subnetName)
	}
	subnetRef := new(vpcref.NewSubnetRef(s.Project, networkName, subnetName))

	virtualMachineName := cmp.Or(s.VirtualMachineName, prefix+"vm")
	ui.Sayf("Creating virtual machine...")
	internalAddress, err := driver.CreateVirtualMachine(ctx, drivermws.CreateVirtualMachineParams{
		VirtualMachineName: virtualMachineName,
		VMType:             s.VMType,
		Zone:               s.Zone,
		SSHUsername:        s.Communicator.SSHUsername,
		SSHPublicKey:       string(s.Communicator.SSHPublicKey),
		CloudConfig:        s.CloudConfig,
		ServiceAccountRef:  s.serviceAccountRef(),
		DiskRef:            diskRef,
		ExternalAddressRef: externalAddressRef,
		SubnetRef:          subnetRef,
	})
	if err != nil {
		return common.ActionHaltWithErrorf(state, "create vm %q: %w", virtualMachineName, err)
	}
	ui.Sayf("Virtual Machine %q created", virtualMachineName)

	if s.UseExternalAddress {
		firewallRuleName := prefix + "ssh-access"
		ui.Sayf("Creating firewall rule...")
		err = driver.CreateFirewallRule(ctx, drivermws.CreateFirewallRuleParams{
			NetworkName:                   networkName,
			FirewallRuleName:              firewallRuleName,
			VirtualMachineInternalAddress: internalAddress.String(),
		})
		if err != nil {
			return common.ActionHaltWithErrorf(state, "create firewall rule %q: %w", firewallRuleName, err)
		}

		ui.Sayf("Firewall Rule %q created", firewallRuleName)
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

	s.GeneratedData.Put("SourceProject", s.SourceProject)
	s.GeneratedData.Put("SourceImageName", s.SourceImage)
	s.GeneratedData.Put("SourceSnapshotName", s.SourceSnapshot)

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

	diskName := cmp.Or(s.DiskName, prefix+"disk")
	common.DeleteWithUI(ctx, ui, "disk", diskName, driver.DeleteDisk)
}

func (s *StepCreateVirtualMachine) imageRef() *computeref.ImageRef {
	if s.SourceImage != "" {
		return new(computeref.NewImageRef(s.SourceProject, s.SourceImage))
	}
	return nil
}

func (s *StepCreateVirtualMachine) snapshotRef() *computeref.SnapshotRef {
	if s.SourceSnapshot != "" {
		return new(computeref.NewSnapshotRef(s.SourceProject, s.SourceSnapshot))
	}
	return nil
}

func (s *StepCreateVirtualMachine) serviceAccountRef() *iamref.ServiceAccountRef {
	if s.VMServiceAccount != "" {
		return new(iamref.NewServiceAccountRef(s.Project, s.VMServiceAccount))
	}
	return nil
}
