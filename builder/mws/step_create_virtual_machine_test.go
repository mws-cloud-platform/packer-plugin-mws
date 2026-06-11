package mws_test

import (
	"cmp"
	"context"
	"errors"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	mockmws "github.com/mws-cloud-platform/packer-plugin-mws/builder/mws/mock"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.uber.org/mock/gomock"
)

func TestStepCreateVirtualMachine_Success(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name   string
		config *mws.Config
	}{
		{
			name: "all_set",
			config: &mws.Config{
				Project:             testProjectName,
				DiskName:            testDiskName,
				NetworkName:         testNetworkName,
				SubnetName:          testSubnetName,
				ExternalAddressName: testExternalAddressName,
				VirtualMachineName:  testVirtualMachineName,
				SourceImage:         testSourceImage,
			},
		},
		{
			name: "network_set",
			config: &mws.Config{
				Project:     testProjectName,
				SourceImage: testSourceImage,
				NetworkName: testNetworkName,
			},
		},
		{
			name: "all_default",
			config: &mws.Config{
				Project:     testProjectName,
				SourceImage: testSourceImage,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			driver := mockmws.NewMockDriver(ctrl)
			writer, state := prepareState(t, tt.config, driver)

			expectedDiskName := cmp.Or(tt.config.DiskName, defaultDiskName)
			expectedExternalAddressName := cmp.Or(tt.config.ExternalAddressName, defaultExternalAddressName)
			expectedNetworkName := cmp.Or(tt.config.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.config.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := cmp.Or(tt.config.VirtualMachineName, defaultVirtualMachineName)
			expectedFirewallRuleName := mws.FirewallRuleName
			expectedDiskRef := new(computeref.NewDiskRef(tt.config.Project, expectedDiskName))

			driver.EXPECT().
				CreateDisk(gomock.Any(), mws.CreateDiskParams{
					DiskName: expectedDiskName,
					DiskType: mws.DefaultDiskType,
					Size:     bytesize.MustParseString(mws.DefaultDiskSize),
					Iops:     mws.DefaultDiskIOPS,
					ImageRef: new(computeref.NewImageRef(tt.config.Project, testSourceImage)),
					Zone:     mws.DefaultZone,
				}).
				Times(1)

			driver.EXPECT().
				CreateExternalAddress(gomock.Any(), mws.CreateExternalAddressParams{
					ExternalAddressName: expectedExternalAddressName,
				}).
				Return(testExternalAddress, nil).
				Times(1)

			if tt.config.NetworkName == "" {
				driver.EXPECT().
					CreateNetwork(gomock.Any(), mws.CreateNetworkParams{
						NetworkName: expectedNetworkName,
					}).
					Times(1)
			}

			if tt.config.SubnetName == "" {
				driver.EXPECT().
					CreateSubnet(gomock.Any(), mws.CreateSubnetParams{
						NetworkName: expectedNetworkName,
						SubnetName:  expectedSubnetName,
						SubnetCidr:  cidraddress.MustParseCIDR4AddressString(mws.DefaultSubnetCidr),
					}).
					Times(1)
			}

			driver.EXPECT().
				CreateVirtualMachine(gomock.Any(), mws.CreateVirtualMachineParams{
					VirtualMachineName: expectedVirtualMachineName,
					VmType:             mws.DefaultVMType,
					Zone:               mws.DefaultZone,
					SSHUsername:        mws.DefaultSSHUsername,
					SSHPublicKey:       testSSHPublicKey,
					DiskRef:            expectedDiskRef,
					ExternalAddressRef: new(vpcref.NewExternalAddressRef(tt.config.Project, expectedExternalAddressName)),
					SubnetRef:          new(vpcref.NewSubnetRef(tt.config.Project, expectedNetworkName, expectedSubnetName)),
				}).
				Return(testInternalAddress, nil).
				Times(1)

			driver.EXPECT().
				CreateFirewallRule(gomock.Any(), mws.CreateFirewallRuleParams{
					NetworkName:                   expectedNetworkName,
					FirewallRuleName:              mws.FirewallRuleName,
					VirtualMachineInternalAddress: testInternalAddress,
				}).
				Times(1)

			driver.EXPECT().DeleteFirewallRule(gomock.Any(), expectedNetworkName, expectedFirewallRuleName).Times(1)
			driver.EXPECT().DeleteVirtualMachine(gomock.Any(), expectedVirtualMachineName).Times(1)
			if tt.config.SubnetName == "" {
				driver.EXPECT().DeleteSubnet(gomock.Any(), expectedNetworkName, expectedSubnetName).Times(1)
			}
			if tt.config.NetworkName == "" {
				driver.EXPECT().DeleteNetwork(gomock.Any(), expectedNetworkName).Times(1)
			}
			driver.EXPECT().DeleteExternalAddress(gomock.Any(), expectedExternalAddressName).Times(1)
			driver.EXPECT().DeleteDisk(gomock.Any(), expectedDiskName).Times(1)

			step := &mws.StepCreateVirtualMachine{
				GeneratedData: &packerbuilderdata.GeneratedData{State: state},
			}

			requireActionContinue(t, state, step.Run(context.Background(), state))
			requireStateGet(t, state, mws.DiskNameKey, expectedDiskName)
			requireStateGet(t, state, mws.ExternalAddressNameKey, expectedExternalAddressName)
			requireStateGet(t, state, mws.NetworkNameKey, expectedNetworkName)
			requireStateGet(t, state, mws.SubnetNameKey, expectedSubnetName)
			requireStateGet(t, state, mws.VirtualMachineNameKey, expectedVirtualMachineName)
			requireStateGet(t, state, mws.FirewallRuleNameKey, expectedFirewallRuleName)
			requireStateGet(t, state, mws.InstanceIpKey, testExternalAddress)
			requireStateGet(t, state, mws.InstanceIdKey, expectedVirtualMachineName)
			requireStateGet(t, state, mws.DiskRefKey, expectedDiskRef)
			requireGeneratedDataGet(t, state, "SourceProject", tt.config.Project)
			requireGeneratedDataGet(t, state, "SourceImageName", testSourceImage)
			step.Cleanup(state)
			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func TestStepCreateVirtualMachine_Error(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name               string
		expectedActionHalt bool
	}{
		{
			name:               "CreateDisk",
			expectedActionHalt: true,
		},
		{
			name:               "CreateExternalAddress",
			expectedActionHalt: true,
		},
		{
			name:               "CreateNetwork",
			expectedActionHalt: true,
		},
		{
			name:               "CreateSubnet",
			expectedActionHalt: true,
		},
		{
			name:               "CreateVirtualMachine",
			expectedActionHalt: true,
		},
		{
			name:               "CreateFirewallRule",
			expectedActionHalt: true,
		},
		{
			name: "DeleteFirewallRule",
		},
		{
			name: "DeleteVirtualMachine",
		},
		{
			name: "DeleteSubnet",
		},
		{
			name: "DeleteNetwork",
		},
		{
			name: "DeleteExternalAddress",
		},
		{
			name: "DeleteDisk",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			driver := mockmws.NewMockDriver(ctrl)

			writer, state := prepareState(t,
				&mws.Config{
					Project:     testProjectName,
					SourceImage: testSourceImage,
				},
				driver,
			)

			expectedErrors := map[string]error{tt.name: errors.New("test error")}
			func() {
				driver.EXPECT().CreateDisk(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateDisk"]).Times(1)
				if tt.name == "CreateDisk" {
					return
				}
				driver.EXPECT().DeleteDisk(gomock.Any(), gomock.Any()).
					Return(expectedErrors["DeleteDisk"]).Times(1)
				driver.EXPECT().CreateExternalAddress(gomock.Any(), gomock.Any()).
					Return(testExternalAddress, expectedErrors["CreateExternalAddress"]).Times(1)
				if tt.name == "CreateExternalAddress" {
					return
				}
				driver.EXPECT().DeleteExternalAddress(gomock.Any(), gomock.Any()).
					Return(expectedErrors["DeleteExternalAddress"]).Times(1)
				driver.EXPECT().CreateNetwork(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateNetwork"]).Times(1)
				if tt.name == "CreateNetwork" {
					return
				}
				driver.EXPECT().DeleteNetwork(gomock.Any(), gomock.Any()).
					Return(expectedErrors["DeleteNetwork"]).Times(1)
				driver.EXPECT().CreateSubnet(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateSubnet"]).Times(1)
				if tt.name == "CreateSubnet" {
					return
				}
				driver.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(expectedErrors["DeleteSubnet"]).Times(1)
				driver.EXPECT().CreateVirtualMachine(gomock.Any(), gomock.Any()).
					Return(testInternalAddress, expectedErrors["CreateVirtualMachine"]).Times(1)
				if tt.name == "CreateVirtualMachine" {
					return
				}
				driver.EXPECT().DeleteVirtualMachine(gomock.Any(), gomock.Any()).
					Return(expectedErrors["DeleteVirtualMachine"]).Times(1)
				driver.EXPECT().CreateFirewallRule(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateFirewallRule"]).Times(1)
				if tt.name == "CreateFirewallRule" {
					return
				}
				driver.EXPECT().DeleteFirewallRule(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(expectedErrors["DeleteFirewallRule"]).Times(1)
			}()

			step := &mws.StepCreateVirtualMachine{
				GeneratedData: &packerbuilderdata.GeneratedData{State: state},
			}

			action := step.Run(context.Background(), state)
			if tt.expectedActionHalt {
				requireActionHalt(t, state, action)
			} else {
				requireActionContinue(t, state, action)
			}
			step.Cleanup(state)
			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}
