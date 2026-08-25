// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"cmp"
	"context"
	"fmt"
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

const serialPort = 1

type StepCreateVirtualMachine struct {
	Project      string
	Zone         string
	Communicator *communicator.Config
	commonconfig.VirtualMachineConfig
	*commonconfig.DiskForExportConfig

	GeneratedData *packerbuilderdata.GeneratedData
	FileWriter    FileWriter
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

	var (
		externalAddress *ipaddress.IPAddress
		internalAddress *ipaddress.IPAddress
	)

	bootDiskSize, err := s.bootDiskSize(ctx, driver)
	if err != nil {
		return common.ActionHaltWithErrorf(state, "calculate boot disk size: %w", err)
	}
	if err = s.createExportDisk(ctx, driver, ui, exportDiskName); err != nil {
		return common.ActionHaltWithError(state, err)
	}
	if err = s.createBootDisk(ctx, driver, ui, bootDiskName, bootDiskSize); err != nil {
		return common.ActionHaltWithError(state, err)
	}
	if externalAddress, err = s.createExternalAddress(ctx, driver, ui, externalAddressName); err != nil {
		return common.ActionHaltWithError(state, err)
	}
	if err = s.createNetwork(ctx, driver, ui, networkName); err != nil {
		return common.ActionHaltWithError(state, err)
	}
	if err = s.createSubnet(ctx, driver, ui, subnetName, networkName); err != nil {
		return common.ActionHaltWithError(state, err)
	}
	if internalAddress, err = s.createVirtualMachine(ctx, driver, ui, virtualMachineName,
		exportDiskName, bootDiskName, externalAddressName, networkName, subnetName); err != nil {
		return common.ActionHaltWithError(state, err)
	}
	if err = s.createFirewallRule(ctx, driver, ui, firewallRuleName, networkName, internalAddress); err != nil {
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
	if s.SourceImage != "" {
		s.GeneratedData.Put("SourceImageName", s.SourceImage)
	} else {
		s.GeneratedData.Put("SourceImageName", s.SourceDiskBackup)
	}
	s.GeneratedData.Put("SourceDiskBackupName", s.SourceDiskBackup)

	return multistep.ActionContinue
}

func (s *StepCreateVirtualMachine) Cleanup(state multistep.StateBag) {
	driver := state.Get(common.DriverKey).(StepCreateVirtualMachineDriver)
	prefix := state.Get(common.PrefixKey).(string)
	ui := state.Get(common.UIKey).(packer.Ui)

	ctx, cancel := context.WithTimeout(context.Background(), s.CleanupTimeout)
	defer cancel()

	exportDiskName := prefix + "disk-for-export"
	bootDiskName := cmp.Or(s.DiskName, prefix+"boot-disk")
	externalAddressName := cmp.Or(s.ExternalAddressName, prefix+"external-address")
	networkName := cmp.Or(s.NetworkName, prefix+"network")
	subnetName := cmp.Or(s.SubnetName, prefix+"subnet")
	virtualMachineName := cmp.Or(s.VirtualMachineName, prefix+"vm")
	firewallRuleName := prefix + "ssh-access"

	s.saveSerialPort(ctx, driver, ui, virtualMachineName)
	if s.UseExternalAddress {
		common.DeleteSubWithUI(ctx, ui, "firewall rule", firewallRuleName, "network", networkName, driver.DeleteFirewallRule)
	}
	common.DeleteWithUI(ctx, ui, "virtual machine", virtualMachineName, driver.DeleteVirtualMachine)
	if s.SubnetName == "" {
		common.DeleteSubWithUI(ctx, ui, "subnet", subnetName, "network", networkName, driver.DeleteSubnet)
	}
	if s.NetworkName == "" {
		common.DeleteWithUI(ctx, ui, "network", networkName, driver.DeleteNetwork)
	}
	if s.UseExternalAddress {
		common.DeleteWithUI(ctx, ui, "external address", externalAddressName, driver.DeleteExternalAddress)
	}
	common.DeleteWithUI(ctx, ui, "boot disk", bootDiskName, driver.DeleteDisk)
	if s.DiskForExportConfig != nil {
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
	// add exportSize second time to provide a safety margin for temporary files
	sum.Add(sum, exportSize.BigInt())
	// convert to GB
	divGB, modGB := new(big.Int).DivMod(sum, bytesize.GB.GetValue(), new(big.Int))
	// round up if necessary
	if modGB.Sign() > 0 {
		divGB.Add(divGB, big.NewInt(1))
	}

	result, err := bytesize.NewFromBigInt(divGB, bytesize.GB)
	return &result, err
}

func (s *StepCreateVirtualMachine) createExportDisk(
	ctx context.Context,
	driver StepCreateVirtualMachineDriver,
	ui packer.Ui,
	exportDiskName string,
) error {
	if s.DiskForExportConfig == nil {
		return nil
	}
	ui.Sayf("Creating export disk %q...", exportDiskName)
	err := driver.CreateDisk(ctx, s.exportDiskParams(exportDiskName))
	if err != nil {
		return fmt.Errorf("create export disk %q: %w", exportDiskName, err)
	}
	ui.Sayf("Export disk %q created", exportDiskName)
	return nil
}

func (s *StepCreateVirtualMachine) createBootDisk(
	ctx context.Context,
	driver StepCreateVirtualMachineDriver,
	ui packer.Ui,
	bootDiskName string,
	bootDiskSize *bytesize.ByteSize,
) error {
	ui.Sayf("Creating boot disk %q...", bootDiskName)
	err := driver.CreateDisk(ctx, drivermws.CreateDiskParams{
		DiskName:      bootDiskName,
		DiskType:      s.DiskType,
		Size:          bootDiskSize,
		Iops:          s.DiskIOPS,
		ImageRef:      s.bootImageRef(),
		DiskBackupRef: s.bootDiskBackupRef(),
		Zone:          s.Zone,
	})
	if err != nil {
		return fmt.Errorf("create boot disk %q: %w", bootDiskName, err)
	}
	ui.Sayf("Boot disk %q created", bootDiskName)
	return nil
}

func (s *StepCreateVirtualMachine) createExternalAddress(
	ctx context.Context,
	driver StepCreateVirtualMachineDriver,
	ui packer.Ui,
	externalAddressName string,
) (*ipaddress.IPAddress, error) {
	if !s.UseExternalAddress {
		return nil, nil
	}
	ui.Sayf("Creating external address %q...", externalAddressName)
	result, err := driver.CreateExternalAddress(ctx, drivermws.CreateExternalAddressParams{
		ExternalAddressName: externalAddressName,
	})
	if err != nil {
		return nil, fmt.Errorf("create external address %q: %w", externalAddressName, err)
	}
	ui.Sayf("External address %q created", externalAddressName)
	return result, nil
}

func (s *StepCreateVirtualMachine) createNetwork(
	ctx context.Context,
	driver StepCreateVirtualMachineDriver,
	ui packer.Ui,
	networkName string,
) error {
	if s.NetworkName != "" {
		return nil
	}
	ui.Sayf("Creating network %q...", networkName)
	err := driver.CreateNetwork(ctx, drivermws.CreateNetworkParams{
		NetworkName: networkName,
	})
	if err != nil {
		return fmt.Errorf("create network %q: %w", networkName, err)
	}
	ui.Sayf("Network %q created", networkName)
	return nil
}

func (s *StepCreateVirtualMachine) createSubnet(
	ctx context.Context,
	driver StepCreateVirtualMachineDriver,
	ui packer.Ui,
	subnetName string,
	networkName string,
) error {
	if s.SubnetName != "" {
		return nil
	}
	ui.Sayf("Creating subnet %q...", subnetName)
	err := driver.CreateSubnet(ctx, drivermws.CreateSubnetParams{
		NetworkName: networkName,
		SubnetName:  subnetName,
		SubnetCidr:  cidraddress.MustParseCIDR4AddressString(s.SubnetCidr),
	})
	if err != nil {
		return fmt.Errorf("create subnet %q: %w", subnetName, err)
	}
	ui.Sayf("Subnet %q created", subnetName)
	return nil
}

func (s *StepCreateVirtualMachine) createVirtualMachine(
	ctx context.Context,
	driver StepCreateVirtualMachineDriver,
	ui packer.Ui,
	virtualMachineName string,
	exportDiskName string,
	bootDiskName string,
	externalAddressName string,
	networkName string,
	subnetName string,
) (*ipaddress.IPAddress, error) {
	ui.Sayf("Creating virtual machine %q...", virtualMachineName)
	result, err := driver.CreateVirtualMachine(ctx, drivermws.CreateVirtualMachineParams{
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
	})
	if err != nil {
		return nil, fmt.Errorf("create virtual machine %q: %w", virtualMachineName, err)
	}
	ui.Sayf("Virtual machine %q created", virtualMachineName)
	return result, nil
}

func (s *StepCreateVirtualMachine) createFirewallRule(
	ctx context.Context,
	driver StepCreateVirtualMachineDriver,
	ui packer.Ui,
	firewallRuleName string,
	networkName string,
	internalAddress *ipaddress.IPAddress,
) error {
	if !s.UseExternalAddress {
		return nil
	}
	ui.Sayf("Creating firewall rule %q...", firewallRuleName)
	err := driver.CreateFirewallRule(ctx, drivermws.CreateFirewallRuleParams{
		NetworkName:                   networkName,
		FirewallRuleName:              firewallRuleName,
		VirtualMachineInternalAddress: internalAddress,
	})
	if err != nil {
		return fmt.Errorf("create firewall rule %q: %w", firewallRuleName, err)
	}
	ui.Sayf("Firewall rule %q created", firewallRuleName)
	return nil
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

// separated from other subStep params because of nil pointer dereference in DiskForExportConfig fields
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

func (s *StepCreateVirtualMachine) saveSerialPort(ctx context.Context, driver StepCreateVirtualMachineDriver, ui packer.Ui, virtualMachineName string) {
	if s.SerialConsoleLogFile == "" {
		return
	}
	ui.Sayf("Getting virtual machine %q serial port output.", virtualMachineName)
	output, err := driver.GetSerialPortOutput(ctx, virtualMachineName, serialPort)
	if err != nil {
		ui.Errorf("Error getting virtual machine %q serial port output.\n"+
			"Error: %v.", virtualMachineName, err)
		return
	}

	fw := cmp.Or[FileWriter](s.FileWriter, RealFileWriter{})
	err = fw.WriteFile(s.SerialConsoleLogFile, output, 0644)
	if err != nil {
		ui.Errorf("Error saving virtual machine %q serial port output to file %q.\n"+
			"Error: %v.", virtualMachineName, s.SerialConsoleLogFile, err)
		return
	}
	ui.Sayf("Virtual machine %q serial port output saved to file %q.", virtualMachineName, s.SerialConsoleLogFile)
}
