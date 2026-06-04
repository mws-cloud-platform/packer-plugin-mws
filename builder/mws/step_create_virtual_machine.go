// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"context"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
)

type StepCreateVirtualMachine struct {
	GeneratedData *packerbuilderdata.GeneratedData
}

func (s *StepCreateVirtualMachine) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	config := state.Get(configKey).(*Config)
	driver := state.Get(driverKey).(Driver)
	prefix := state.Get(uuidPrefixKey).(string)
	ui := state.Get(uiKey).(packer.Ui)

	var (
		imageRef    *computeref.ImageRef
		snapshotRef *computeref.SnapshotRef
	)

	if config.SourceImage != "" {
		imageRef = new(computeref.NewImageRef(config.SourceProject, config.SourceImage))
	}
	if config.SourceSnapshot != "" {
		snapshotRef = new(computeref.NewSnapshotRef(config.SourceProject, config.SourceSnapshot))
	}

	diskName := config.DiskName
	if diskName == "" {
		diskName = prefix + "disk"
	}
	ui.Sayf("Creating disk...")
	err := driver.CreateDisk(ctx, CreateDiskParams{
		DiskName:    diskName,
		DiskType:    config.DiskType,
		Size:        bytesize.MustParseString(config.DiskSize),
		Iops:        config.IOPS,
		ImageRef:    imageRef,
		SnapshotRef: snapshotRef,
		Zone:        config.Zone,
	})
	if err != nil {
		return actionHaltWithError(state, err)
	}

	ui.Sayf("Disk %q created", diskName)
	state.Put(diskNameKey, diskName)

	diskRef := new(computeref.NewDiskRef(config.Project, diskName))
	state.Put(diskRefKey, diskRef)

	externalAddressName := config.ExternalAddressName
	if externalAddressName == "" {
		externalAddressName = prefix + "external-address"
	}
	ui.Sayf("Creating external address...")
	externalAddress, err := driver.CreateExternalAddress(ctx, CreateExternalAddressParams{
		ExternalAddressName: externalAddressName,
	})
	if err != nil {
		return actionHaltWithError(state, err)
	}

	ui.Sayf("External Address %q created", externalAddressName)
	state.Put(externalAddressNameKey, externalAddressName)
	state.Put(instanceIpKey, externalAddress)
	externalAddressRef := new(vpcref.NewExternalAddressRef(config.Project, externalAddressName))

	networkName := config.NetworkName
	if config.NetworkName == "" {
		networkName = prefix + "network"
		ui.Sayf("Creating network...")
		err = driver.CreateNetwork(ctx, CreateNetworkParams{
			NetworkName: networkName,
		})
		if err != nil {
			return actionHaltWithError(state, err)
		}

		ui.Sayf("Network %q created", networkName)
	}
	state.Put(networkNameKey, networkName)

	subnetName := config.SubnetName
	if subnetName == "" {
		subnetName = prefix + "subnet"
		ui.Sayf("Creating subnet...")
		err = driver.CreateSubnet(ctx, CreateSubnetParams{
			NetworkName: networkName,
			SubnetName:  subnetName,
			SubnetCidr:  cidraddress.MustParseCIDR4AddressString(config.SubnetCidr),
		})
		if err != nil {
			return actionHaltWithError(state, err)
		}

		ui.Sayf("Subnet %q created", subnetName)
	}
	state.Put(subnetNameKey, subnetName)
	subnetRef := new(vpcref.NewSubnetRef(config.Project, networkName, subnetName))

	virtualMachineName := config.VirtualMachineName
	if virtualMachineName == "" {
		virtualMachineName = prefix + "vm"
	}
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
		return actionHaltWithError(state, err)
	}

	ui.Sayf("Virtual Machine %q created", virtualMachineName)
	state.Put(virtualMachineNameKey, virtualMachineName)

	firewallRuleName := "access-from-internet-ssh"
	ui.Sayf("Creating firewall rule...")
	err = driver.CreateFirewallRule(ctx, CreateFirewallRuleParams{
		NetworkName:                   networkName,
		FirewallRuleName:              firewallRuleName,
		VirtualMachineInternalAddress: internalAddress,
	})
	if err != nil {
		return actionHaltWithError(state, err)
	}

	ui.Sayf("Firewall Rule %q created", firewallRuleName)
	state.Put(firewallRuleNameKey, firewallRuleName)

	// instance_id is the generic term used so that users can have access to the
	// instance id inside of the provisioners, used in step_provision.
	state.Put(instanceIdKey, virtualMachineName)

	s.GeneratedData.Put("SourceProject", config.SourceProject)
	s.GeneratedData.Put("SourceImageName", config.SourceImage)
	s.GeneratedData.Put("SourceSnapshotName", config.SourceSnapshot)

	return multistep.ActionContinue
}

func (s *StepCreateVirtualMachine) Cleanup(state multistep.StateBag) {
	config := state.Get(configKey).(*Config)
	driver := state.Get(driverKey).(Driver)
	ui := state.Get(uiKey).(packer.Ui)

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
	if name, ok := state.GetOk(diskNameKey); ok {
		diskName = name.(string)
	}
	if name, ok := state.GetOk(externalAddressNameKey); ok {
		externalAddressName = name.(string)
	}
	if name, ok := state.GetOk(networkNameKey); ok {
		networkName = name.(string)
	}
	if name, ok := state.GetOk(subnetNameKey); ok {
		subnetName = name.(string)
	}
	if name, ok := state.GetOk(virtualMachineNameKey); ok {
		virtualMachineName = name.(string)
	}
	if name, ok := state.GetOk(firewallRuleNameKey); ok {
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
