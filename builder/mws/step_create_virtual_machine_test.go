package mws_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	mock_mws "github.com/mws-cloud-platform/packer-plugin-mws/builder/mws/mock"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
	"go.uber.org/mock/gomock"
)

func TestStepCreateVirtualMachine_Run_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	diskRef := new(computeref.NewDiskRef(testProjectName, testDiskName))
	externalAddressRef := new(vpcref.NewExternalAddressRef(testProjectName, testExternalAddr))
	subnetRef := new(vpcref.NewSubnetRef(testProjectName, testNetworkName, testSubnetName))

	driver := mock_mws.NewMockDriver(ctrl)

	driver.EXPECT().
		CreateDisk(gomock.Any(), mws.CreateDiskParams{
			DiskName: testDiskName,
			DiskType: mws.DefaultDiskType,
			Size:     bytesize.MustParseString(mws.DefaultDiskSize),
			Iops:     mws.DefaultDiskIOPS,
			ImageRef: new(computeref.NewImageRef(testProjectName, testSourceImage)),
			Zone:     mws.DefaultZone,
		}).
		Return(nil).
		Times(1)

	driver.EXPECT().
		CreateExternalAddress(gomock.Any(), mws.CreateExternalAddressParams{
			ExternalAddressName: testExternalAddr,
		}).
		Return(testExternalAddress, nil).
		Times(1)

	driver.EXPECT().
		CreateVirtualMachine(gomock.Any(), mws.CreateVirtualMachineParams{
			VirtualMachineName: testVmName,
			VmType:             mws.DefaultVMType,
			Zone:               mws.DefaultZone,
			SSHUsername:        mws.DefaultSSHUsername,
			SSHPublicKey:       testSSHPublicKey,
			DiskRef:            diskRef,
			ExternalAddressRef: externalAddressRef,
			SubnetRef:          subnetRef,
		}).
		Return(testInternalAddress, nil).
		Times(1)

	driver.EXPECT().
		CreateFirewallRule(gomock.Any(), mws.CreateFirewallRuleParams{
			NetworkName:                   testNetworkName,
			FirewallRuleName:              mws.FirewallRuleName,
			VirtualMachineInternalAddress: testInternalAddress,
		}).
		Return(nil).
		Times(1)

	config := &mws.Config{
		Project:             testProjectName,
		DiskName:            testDiskName,
		NetworkName:         testNetworkName,
		SubnetName:          testSubnetName,
		ExternalAddressName: testExternalAddr,
		VirtualMachineName:  testVmName,
		SourceImage:         testSourceImage,
	}
	config.SetDefaults()
	require.NoError(t, config.Validate())
	config.Communicator.SSHPublicKey = []byte(testSSHPublicKey)

	state := new(multistep.BasicStateBag)
	state.Put(mws.ConfigKey, config)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UuidPrefixKey, packerPrefix)

	var writer bytes.Buffer
	ui := &packer.BasicUi{
		Writer: &writer,
	}
	state.Put(mws.UiKey, ui)

	step := &mws.StepCreateVirtualMachine{
		GeneratedData: &packerbuilderdata.GeneratedData{State: state},
	}

	requireActionContinue(t, state, step.Run(context.Background(), state))
	requireStateGet(t, state, mws.DiskNameKey, testDiskName)
	requireStateGet(t, state, mws.ExternalAddressNameKey, testExternalAddr)
	requireStateGet(t, state, mws.NetworkNameKey, testNetworkName)
	requireStateGet(t, state, mws.SubnetNameKey, testSubnetName)
	requireStateGet(t, state, mws.VirtualMachineNameKey, testVmName)
	requireStateGet(t, state, mws.FirewallRuleNameKey, mws.FirewallRuleName)
	requireStateGet(t, state, mws.InstanceIpKey, testExternalAddress)
	requireStateGet(t, state, mws.InstanceIdKey, testVmName)
	requireStateGet(t, state, mws.DiskRefKey, diskRef)
	requireGeneratedDataGet(t, state, "SourceProject", testProjectName)
	requireGeneratedDataGet(t, state, "SourceImageName", testSourceImage)
	requireOutput(t, writer.String())
}

func TestStepCreateVirtualMachine_Run_WithDefaultNames(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	diskRef := new(computeref.NewDiskRef(testProjectName, defaultDiskName))
	externalAddressRef := new(vpcref.NewExternalAddressRef(testProjectName, defaultExternalAddressName))
	subnetRef := new(vpcref.NewSubnetRef(testProjectName, defaultNetworkName, defaultSubnetName))

	driver := mock_mws.NewMockDriver(ctrl)

	driver.EXPECT().
		CreateDisk(gomock.Any(), mws.CreateDiskParams{
			DiskName: defaultDiskName,
			DiskType: mws.DefaultDiskType,
			Size:     bytesize.MustParseString(mws.DefaultDiskSize),
			Iops:     mws.DefaultDiskIOPS,
			ImageRef: new(computeref.NewImageRef(testProjectName, testSourceImage)),
			Zone:     mws.DefaultZone,
		}).
		Return(nil).
		Times(1)

	driver.EXPECT().
		CreateExternalAddress(gomock.Any(), mws.CreateExternalAddressParams{
			ExternalAddressName: defaultExternalAddressName,
		}).
		Return(testExternalAddress, nil).
		Times(1)

	driver.EXPECT().
		CreateNetwork(gomock.Any(), mws.CreateNetworkParams{
			NetworkName: defaultNetworkName,
		}).
		Return(nil).
		Times(1)

	driver.EXPECT().
		CreateSubnet(gomock.Any(), mws.CreateSubnetParams{
			NetworkName: defaultNetworkName,
			SubnetName:  defaultSubnetName,
			SubnetCidr:  cidraddress.MustParseCIDR4AddressString(mws.DefaultSubnetCidr),
		}).
		Return(nil).
		Times(1)

	driver.EXPECT().
		CreateVirtualMachine(gomock.Any(), mws.CreateVirtualMachineParams{
			VirtualMachineName: defaultVmName,
			VmType:             mws.DefaultVMType,
			Zone:               mws.DefaultZone,
			SSHUsername:        mws.DefaultSSHUsername,
			SSHPublicKey:       testSSHPublicKey,
			DiskRef:            diskRef,
			ExternalAddressRef: externalAddressRef,
			SubnetRef:          subnetRef,
		}).
		Return(testInternalAddress, nil).
		Times(1)

	driver.EXPECT().
		CreateFirewallRule(gomock.Any(), mws.CreateFirewallRuleParams{
			NetworkName:                   defaultNetworkName,
			FirewallRuleName:              mws.FirewallRuleName,
			VirtualMachineInternalAddress: testInternalAddress,
		}).
		Return(nil).
		Times(1)

	config := &mws.Config{
		Project:     testProjectName,
		SourceImage: testSourceImage,
	}
	config.SetDefaults()
	require.NoError(t, config.Validate())
	config.Communicator.SSHPublicKey = []byte(testSSHPublicKey)

	state := new(multistep.BasicStateBag)
	state.Put(mws.ConfigKey, config)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UuidPrefixKey, packerPrefix)

	var writer bytes.Buffer
	ui := &packer.BasicUi{
		Writer: &writer,
	}
	state.Put(mws.UiKey, ui)

	step := &mws.StepCreateVirtualMachine{
		GeneratedData: &packerbuilderdata.GeneratedData{State: state},
	}

	requireActionContinue(t, state, step.Run(context.Background(), state))
	requireStateGet(t, state, mws.DiskNameKey, defaultDiskName)
	requireStateGet(t, state, mws.ExternalAddressNameKey, defaultExternalAddressName)
	requireStateGet(t, state, mws.NetworkNameKey, defaultNetworkName)
	requireStateGet(t, state, mws.SubnetNameKey, defaultSubnetName)
	requireStateGet(t, state, mws.VirtualMachineNameKey, defaultVmName)
	requireStateGet(t, state, mws.FirewallRuleNameKey, mws.FirewallRuleName)
	requireStateGet(t, state, mws.InstanceIpKey, testExternalAddress)
	requireStateGet(t, state, mws.InstanceIdKey, defaultVmName)
	requireStateGet(t, state, mws.DiskRefKey, diskRef)
	requireGeneratedDataGet(t, state, "SourceProject", testProjectName)
	requireGeneratedDataGet(t, state, "SourceImageName", testSourceImage)
	requireOutput(t, writer.String())
}

func TestStepCreateVirtualMachine_Run_CreateDiskError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	driver := mock_mws.NewMockDriver(ctrl)
	driver.EXPECT().
		CreateDisk(gomock.Any(), gomock.Any()).
		Return(errors.New("disk creation failed")).
		Times(1)

	config := &mws.Config{
		Project:     testProjectName,
		SourceImage: testSourceImage,
	}
	config.SetDefaults()
	require.NoError(t, config.Validate())

	state := new(multistep.BasicStateBag)
	state.Put(mws.ConfigKey, config)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UuidPrefixKey, packerPrefix)

	var writer bytes.Buffer
	ui := &packer.BasicUi{
		Writer: &writer,
	}
	state.Put(mws.UiKey, ui)

	step := &mws.StepCreateVirtualMachine{
		GeneratedData: &packerbuilderdata.GeneratedData{State: state},
	}

	requireActionHalt(t, state, step.Run(context.Background(), state))
	requireOutput(t, writer.String())
}
