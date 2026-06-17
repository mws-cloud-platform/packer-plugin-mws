// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"cmp"
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
)

const (
	FirewallRuleName = "access-from-internet-ssh"
)

type StepCreateVirtualMachine struct {
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateVirtualMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(ConfigKey).(*Config)
	driver := state.Get(DriverKey).(Driver)
	prefix := state.Get(UuidPrefixKey).(string)
	ui := state.Get(UiKey).(packer.Ui)

	var (
		imageRef           *computeref.ImageRef
		snapshotRef        *computeref.SnapshotRef
		externalAddressRef *vpcref.ExternalAddressRef
	)

	if config.SourceImage != "" {
		imageRef = new(computeref.NewImageRef(config.SourceProject, config.SourceImage))
	}
	if config.SourceSnapshot != "" {
		snapshotRef = new(computeref.NewSnapshotRef(config.SourceProject, config.SourceSnapshot))
	}

	diskName := cmp.Or(config.DiskName, prefix+"disk")
	ui.Sayf("Creating disk...")
	err := driver.CreateDisk(ctx, CreateDiskParams{
		DiskName:    diskName,
		DiskType:    config.DiskType,
		Size:        bytesize.MustParseString(config.DiskSize),
		Iops:        config.DiskIOPS,
		ImageRef:    imageRef,
		SnapshotRef: snapshotRef,
		Zone:        config.Zone,
	})
	if err != nil {
		return actionHaltWithError(state, fmt.Errorf("create disk %q: %w", diskName, err))
	}

	ui.Sayf("Disk %q created", diskName)
	state.Put(DiskNameKey, diskName)

	diskRef := new(computeref.NewDiskRef(config.Project, diskName))
	state.Put(DiskRefKey, diskRef)

	if config.UseExternalAddress {
		externalAddressName := cmp.Or(config.ExternalAddressName, prefix+"external-address")
		ui.Sayf("Creating external address...")
		externalAddress, err := driver.CreateExternalAddress(ctx, CreateExternalAddressParams{
			ExternalAddressName: externalAddressName,
		})
		if err != nil {
			return actionHaltWithError(state, fmt.Errorf("create external-address %q: %w", externalAddressName, err))
		}

		ui.Sayf("External Address %q created", externalAddressName)
		state.Put(ExternalAddressNameKey, externalAddressName)
		state.Put(InstanceIpKey, externalAddress)
		externalAddressRef = new(vpcref.NewExternalAddressRef(config.Project, externalAddressName))
	}

	networkName := cmp.Or(config.NetworkName, prefix+"network")
	if config.NetworkName == "" {
		ui.Sayf("Creating network...")
		err = driver.CreateNetwork(ctx, CreateNetworkParams{
			NetworkName: networkName,
		})
		if err != nil {
			return actionHaltWithError(state, fmt.Errorf("create network %q: %w", networkName, err))
		}

		ui.Sayf("Network %q created", networkName)
	}
	state.Put(NetworkNameKey, networkName)

	subnetName := cmp.Or(config.SubnetName, prefix+"subnet")
	if config.SubnetName == "" {
		ui.Sayf("Creating subnet...")
		err = driver.CreateSubnet(ctx, CreateSubnetParams{
			NetworkName: networkName,
			SubnetName:  subnetName,
			SubnetCidr:  cidraddress.MustParseCIDR4AddressString(config.SubnetCidr),
		})
		if err != nil {
			return actionHaltWithError(state, fmt.Errorf("create subnet %q: %w", subnetName, err))
		}

		ui.Sayf("Subnet %q created", subnetName)
	}
	state.Put(SubnetNameKey, subnetName)
	subnetRef := new(vpcref.NewSubnetRef(config.Project, networkName, subnetName))

	virtualMachineName := cmp.Or(config.VirtualMachineName, prefix+"vm")
	ui.Sayf("Creating virtual machine...")
	internalAddress, err := driver.CreateVirtualMachine(ctx, CreateVirtualMachineParams{
		VirtualMachineName: virtualMachineName,
		VmType:             config.VmType,
		Zone:               config.Zone,
		SSHUsername:        config.Communicator.SSHUsername,
		SSHPublicKey:       string(config.Communicator.SSHPublicKey),
		DiskRef:            diskRef,
		ExternalAddressRef: externalAddressRef,
		SubnetRef:          subnetRef,
	})
	if err != nil {
		return actionHaltWithError(state, fmt.Errorf("create vm %q: %w", virtualMachineName, err))
	}

	ui.Sayf("Virtual Machine %q created", virtualMachineName)
	state.Put(VirtualMachineNameKey, virtualMachineName)

	if config.UseExternalAddress {
		ui.Sayf("Creating firewall rule...")
		err = driver.CreateFirewallRule(ctx, CreateFirewallRuleParams{
			NetworkName:                   networkName,
			FirewallRuleName:              FirewallRuleName,
			VirtualMachineInternalAddress: internalAddress,
		})
		if err != nil {
			return actionHaltWithError(state, fmt.Errorf("create firewall rule %q: %w", FirewallRuleName, err))
		}

		ui.Sayf("Firewall Rule %q created", FirewallRuleName)
		state.Put(FirewallRuleNameKey, FirewallRuleName)
	} else {
		state.Put(InstanceIpKey, internalAddress)
	}

	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put(InstanceIdKey, virtualMachineName)

	s.GeneratedData.Put("SourceProject", config.SourceProject)
	s.GeneratedData.Put("SourceImageName", config.SourceImage)
	s.GeneratedData.Put("SourceSnapshotName", config.SourceSnapshot)

	return multistep.ActionContinue
}

func (s *StepCreateVirtualMachine) Cleanup(state multistep.StateBag) {
	config := state.Get(ConfigKey).(*Config)
	driver := state.Get(DriverKey).(Driver)
	ui := state.Get(UiKey).(packer.Ui)

	cleanupTimeout, _ := time.ParseDuration(config.CleanupTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), cleanupTimeout)
	defer cancel()

	var (
		diskName            string
		externalAddressName string
		networkName         string
		subnetName          string
		virtualMachineName  string
		firewallRuleName    string
	)
	if name, ok := state.GetOk(DiskNameKey); ok {
		diskName = name.(string)
	}
	if name, ok := state.GetOk(ExternalAddressNameKey); ok {
		externalAddressName = name.(string)
	}
	if name, ok := state.GetOk(NetworkNameKey); ok {
		networkName = name.(string)
	}
	if name, ok := state.GetOk(SubnetNameKey); ok {
		subnetName = name.(string)
	}
	if name, ok := state.GetOk(VirtualMachineNameKey); ok {
		virtualMachineName = name.(string)
	}
	if name, ok := state.GetOk(FirewallRuleNameKey); ok {
		firewallRuleName = name.(string)
	}

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

	if subnetName != "" && config.SubnetName == "" {
		if err := driver.DeleteSubnet(ctx, networkName, subnetName); err != nil {
			ui.Errorf("Error deleting subnet %q in network %q. Please delete it manually.\n"+
				"Error: %v.", subnetName, networkName, err)
		} else {
			ui.Sayf("Subnet %q deleted", subnetName)
		}
	}

	if networkName != "" && config.NetworkName == "" {
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
