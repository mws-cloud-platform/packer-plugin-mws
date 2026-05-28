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
	Debug         bool
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
	if s.Debug {
		ui.Sayf("Disk created %s", diskName)
	}
	state.Put(diskNameKey, diskName)

	diskRef := new(computeref.NewDiskRef(config.Project, diskName))
	state.Put(diskRefKey, diskRef)

	externalAddressName := config.ExternalAddressName
	if externalAddressName == "" {
		externalAddressName = prefix + "external-address"
	}
	externalAddress, err := driver.CreateExternalAddress(ctx, CreateExternalAddressParams{
		ExternalAddressName: externalAddressName,
	})
	if err != nil {
		return actionHaltWithError(state, err)
	}
	if s.Debug {
		ui.Sayf("External Address created %s", externalAddressName)
	}
	state.Put(externalAddressNameKey, externalAddressName)
	state.Put(instanceIpKey, externalAddress)
	externalAddressRef := new(vpcref.NewExternalAddressRef(config.Project, externalAddressName))

	networkName := config.NetworkName
	if config.NetworkName == "" {
		networkName = prefix + "network"
		err = driver.CreateNetwork(ctx, CreateNetworkParams{
			NetworkName: networkName,
		})
		if err != nil {
			return actionHaltWithError(state, err)
		}
		if s.Debug {
			ui.Sayf("Network created %s", networkName)
		}
	}
	state.Put(networkNameKey, networkName)

	subnetName := config.SubnetName
	if subnetName == "" {
		subnetName = prefix + "subnet"
		err = driver.CreateSubnet(ctx, CreateSubnetParams{
			NetworkName: networkName,
			SubnetName:  subnetName,
			SubnetCidr:  cidraddress.MustParseCIDR4AddressString(config.SubnetCidr),
		})
		if err != nil {
			return actionHaltWithError(state, err)
		}
		if s.Debug {
			ui.Sayf("Subnet created %s", subnetName)
		}
	}
	state.Put(subnetNameKey, subnetName)
	subnetRef := new(vpcref.NewSubnetRef(config.Project, networkName, subnetName))

	virtualMachineName := config.VirtualMachineName
	if virtualMachineName == "" {
		virtualMachineName = prefix + "vm"
	}
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
	if s.Debug {
		ui.Sayf("Virtual Machine created %s", virtualMachineName)
	}
	state.Put(virtualMachineNameKey, virtualMachineName)

	firewallRuleName := "access-from-internet-ssh"
	err = driver.CreateFirewallRule(ctx, CreateFirewallRuleParams{
		NetworkName:                   networkName,
		FirewallRuleName:              firewallRuleName,
		VirtualMachineInternalAddress: internalAddress,
	})
	if err != nil {
		return actionHaltWithError(state, err)
	}
	if s.Debug {
		ui.Sayf("Firewall Rule created %s", firewallRuleName)
	}
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
			ui.Errorf("Error deleting firewall rule %s in network %s: %v", firewallRuleName, networkName, err)
		} else if s.Debug {
			ui.Sayf("Firewall Rule deleted %s", firewallRuleName)
		}
	}

	if virtualMachineName != "" {
		if err := driver.DeleteVirtualMachine(ctx, virtualMachineName); err != nil {
			ui.Errorf("Error deleting virtual machine %s: %v", virtualMachineName, err)
		} else if s.Debug {
			ui.Sayf("Virtual Machine deleted %s", virtualMachineName)
		}
	}

	if subnetName != "" && config.SubnetName == "" {
		if err := driver.DeleteSubnet(ctx, networkName, subnetName); err != nil {
			ui.Errorf("Error deleting subnet %s in network %s: %v", subnetName, networkName, err)
		} else if s.Debug {
			ui.Sayf("Subnet deleted %s", subnetName)
		}
	}

	if networkName != "" && config.NetworkName == "" {
		if err := driver.DeleteNetwork(ctx, networkName); err != nil {
			ui.Errorf("Error deleting network %s: %v", networkName, err)
		} else if s.Debug {
			ui.Sayf("Network deleted %s", networkName)
		}
	}

	if externalAddressName != "" {
		if err := driver.DeleteExternalAddress(ctx, externalAddressName); err != nil {
			ui.Errorf("Error deleting external address %s: %v", externalAddressName, err)
		} else if s.Debug {
			ui.Sayf("External address deleted %s", externalAddressName)
		}
	}

	if diskName != "" {
		if err := driver.DeleteDisk(ctx, diskName); err != nil {
			ui.Errorf("Error deleting disk %s: %v", diskName, err)
		} else if s.Debug {
			ui.Sayf("Disk deleted %s", diskName)
		}
	}
}
