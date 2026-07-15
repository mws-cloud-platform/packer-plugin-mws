// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"cmp"
	"context"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	commonconfig "github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
)

const (
	FirewallRuleName = "access-from-internet-ssh"
)

type StepCreateVirtualMachine struct {
	Communicator communicator.Config `mapstructure:",squash" json:"-"`
	commonconfig.AccessConfig
	commonconfig.VirtualMachineConfig

	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateVirtualMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(DriverKey).(StepCreateVirtualMachineDriver)
	prefix := state.Get(PrefixKey).(string)
	ui := state.Get(UIKey).(packer.Ui)

	var (
		imageRef           *computeref.ImageRef
		snapshotRef        *computeref.SnapshotRef
		externalAddressRef *vpcref.ExternalAddressRef
	)

	if s.SourceImage != "" {
		imageRef = new(computeref.NewImageRef(s.SourceProject, s.SourceImage))
	}
	if s.SourceSnapshot != "" {
		snapshotRef = new(computeref.NewSnapshotRef(s.SourceProject, s.SourceSnapshot))
	}

	diskName := cmp.Or(s.DiskName, prefix+"disk")
	ui.Sayf("Creating disk...")
	if err := driver.CreateDisk(ctx, drivermws.CreateDiskParams{
		DiskName:    diskName,
		DiskType:    s.DiskType,
		Size:        bytesize.MustParseString(s.DiskSize),
		Iops:        s.DiskIOPS,
		ImageRef:    imageRef,
		SnapshotRef: snapshotRef,
		Zone:        s.Zone,
	}); err != nil {
		return ActionHaltWithErrorf(state, "create disk %q: %w", diskName, err)
	}

	ui.Sayf("Disk %q created", diskName)
	state.Put(DiskNameKey, diskName)

	diskRef := new(computeref.NewDiskRef(s.Project, diskName))
	state.Put(DiskRefKey, diskRef)

	if s.UseExternalAddress {
		externalAddressName := cmp.Or(s.ExternalAddressName, prefix+"external-address")
		ui.Sayf("Creating external address...")
		externalAddress, err := driver.CreateExternalAddress(ctx, drivermws.CreateExternalAddressParams{
			ExternalAddressName: externalAddressName,
		})
		if err != nil {
			return ActionHaltWithErrorf(state, "create external-address %q: %w", externalAddressName, err)
		}

		ui.Sayf("External Address %q created", externalAddressName)
		state.Put(ExternalAddressNameKey, externalAddressName)
		state.Put(InstanceIPKey, externalAddress)
		externalAddressRef = new(vpcref.NewExternalAddressRef(s.Project, externalAddressName))
	}

	networkName := cmp.Or(s.NetworkName, prefix+"network")
	if s.NetworkName == "" {
		ui.Sayf("Creating network...")
		if err := driver.CreateNetwork(ctx, drivermws.CreateNetworkParams{
			NetworkName: networkName,
		}); err != nil {
			return ActionHaltWithErrorf(state, "create network %q: %w", networkName, err)
		}

		ui.Sayf("Network %q created", networkName)
	}
	state.Put(NetworkNameKey, networkName)

	subnetName := cmp.Or(s.SubnetName, prefix+"subnet")
	if s.SubnetName == "" {
		ui.Sayf("Creating subnet...")
		if err := driver.CreateSubnet(ctx, drivermws.CreateSubnetParams{
			NetworkName: networkName,
			SubnetName:  subnetName,
			SubnetCidr:  cidraddress.MustParseCIDR4AddressString(s.SubnetCidr),
		}); err != nil {
			return ActionHaltWithErrorf(state, "create subnet %q: %w", subnetName, err)
		}

		ui.Sayf("Subnet %q created", subnetName)
	}
	state.Put(SubnetNameKey, subnetName)
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
		DiskRef:            diskRef,
		ExternalAddressRef: externalAddressRef,
		SubnetRef:          subnetRef,
	})
	if err != nil {
		return ActionHaltWithErrorf(state, "create vm %q: %w", virtualMachineName, err)
	}

	ui.Sayf("Virtual Machine %q created", virtualMachineName)
	state.Put(VirtualMachineNameKey, virtualMachineName)

	if s.UseExternalAddress {
		ui.Sayf("Creating firewall rule...")
		err = driver.CreateFirewallRule(ctx, drivermws.CreateFirewallRuleParams{
			NetworkName:                   networkName,
			FirewallRuleName:              FirewallRuleName,
			VirtualMachineInternalAddress: internalAddress,
		})
		if err != nil {
			return ActionHaltWithErrorf(state, "create firewall rule %q: %w", FirewallRuleName, err)
		}

		ui.Sayf("Firewall Rule %q created", FirewallRuleName)
		state.Put(FirewallRuleNameKey, FirewallRuleName)
	} else {
		state.Put(InstanceIPKey, internalAddress)
	}

	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put(InstanceIDKey, virtualMachineName)

	s.GeneratedData.Put("SourceProject", s.SourceProject)
	s.GeneratedData.Put("SourceImageName", s.SourceImage)
	s.GeneratedData.Put("SourceSnapshotName", s.SourceSnapshot)

	return multistep.ActionContinue
}

func (s *StepCreateVirtualMachine) Cleanup(state multistep.StateBag) {
	driver := state.Get(DriverKey).(StepCreateVirtualMachineDriver)
	ui := state.Get(UIKey).(packer.Ui)

	ctx, cancel := context.WithTimeout(context.Background(), s.CleanupTimeout)
	defer cancel()

	diskName := stateGetOkString(state, DiskNameKey)
	externalAddressName := stateGetOkString(state, ExternalAddressNameKey)
	networkName := stateGetOkString(state, NetworkNameKey)
	subnetName := stateGetOkString(state, SubnetNameKey)
	virtualMachineName := stateGetOkString(state, VirtualMachineNameKey)
	firewallRuleName := stateGetOkString(state, FirewallRuleNameKey)

	if firewallRuleName != "" {
		if err := driver.DeleteFirewallRule(ctx, networkName, firewallRuleName); err != nil {
			ui.Errorf("Error deleting firewall rule %q in network %q. Please delete it manually.\n"+
				"Error: %v.", firewallRuleName, networkName, err)
		} else {
			ui.Sayf("Firewall Rule %q deleted", firewallRuleName)
		}
	}

	if virtualMachineName != "" {
		if err := driver.DeleteVirtualMachine(ctx, virtualMachineName); err != nil {
			ui.Errorf("Error deleting virtual machine %q. Please delete it manually.\n"+
				"Error: %v.", virtualMachineName, err)
		} else {
			ui.Sayf("Virtual Machine %q deleted", virtualMachineName)
		}
	}

	if subnetName != "" && s.SubnetName == "" {
		if err := driver.DeleteSubnet(ctx, networkName, subnetName); err != nil {
			ui.Errorf("Error deleting subnet %q in network %q. Please delete it manually.\n"+
				"Error: %v.", subnetName, networkName, err)
		} else {
			ui.Sayf("Subnet %q deleted", subnetName)
		}
	}

	if networkName != "" && s.NetworkName == "" {
		if err := driver.DeleteNetwork(ctx, networkName); err != nil {
			ui.Errorf("Error deleting network %q. Please delete it manually.\n"+
				"Error: %v.", networkName, err)
		} else {
			ui.Sayf("Network %q deleted", networkName)
		}
	}

	if externalAddressName != "" {
		if err := driver.DeleteExternalAddress(ctx, externalAddressName); err != nil {
			ui.Errorf("Error deleting external address %q. Please delete it manually.\n"+
				"Error: %v.", externalAddressName, err)
		} else {
			ui.Sayf("External address %q deleted", externalAddressName)
		}
	}

	if diskName != "" {
		if err := driver.DeleteDisk(ctx, diskName); err != nil {
			ui.Errorf("Error deleting disk %q. Please delete it manually.\n"+
				"Error: %v.", diskName, err)
		} else {
			ui.Sayf("Disk %q deleted", diskName)
		}
	}
}
