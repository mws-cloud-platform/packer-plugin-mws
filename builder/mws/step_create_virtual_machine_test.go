// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"cmp"
	"context"
	"errors"
	"fmt"
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

func TestStepCreateVirtualMachine_Run_Success(t *testing.T) {
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
			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func TestStepCreateVirtualMachine_Cleanup_Success(t *testing.T) {
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
		for _, errorInRun := range []string{
			"CreateDisk",
			"CreateExternalAddress",
			"CreateNetwork",
			"CreateSubnet",
			"CreateVirtualMachine",
			"CreateFirewallRule",
			"None",
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

				func() {
					if errorInRun == "CreateDisk" {
						return
					}
					state.Put(mws.DiskNameKey, expectedDiskName)
					driver.EXPECT().DeleteDisk(gomock.Any(), expectedDiskName).Times(1)
					if errorInRun == "CreateExternalAddress" {
						return
					}
					state.Put(mws.ExternalAddressNameKey, expectedExternalAddressName)
					driver.EXPECT().DeleteExternalAddress(gomock.Any(), expectedExternalAddressName).Times(1)
					if errorInRun == "CreateNetwork" {
						return
					}
					state.Put(mws.NetworkNameKey, expectedNetworkName)
					if tt.config.NetworkName == "" {
						driver.EXPECT().DeleteNetwork(gomock.Any(), expectedNetworkName).Times(1)
					}
					if errorInRun == "CreateSubnet" {
						return
					}
					state.Put(mws.SubnetNameKey, expectedSubnetName)
					if tt.config.SubnetName == "" {
						driver.EXPECT().DeleteSubnet(gomock.Any(), expectedNetworkName, expectedSubnetName).Times(1)
					}
					if errorInRun == "CreateVirtualMachine" {
						return
					}
					state.Put(mws.VirtualMachineNameKey, expectedVirtualMachineName)
					driver.EXPECT().DeleteVirtualMachine(gomock.Any(), expectedVirtualMachineName).Times(1)
					if errorInRun == "CreateFirewallRule" {
						return
					}
					state.Put(mws.FirewallRuleNameKey, expectedFirewallRuleName)
					driver.EXPECT().DeleteFirewallRule(gomock.Any(), expectedNetworkName, expectedFirewallRuleName).Times(1)
				}()

				step := &mws.StepCreateVirtualMachine{
					GeneratedData: &packerbuilderdata.GeneratedData{State: state},
				}

				step.Cleanup(state)
				expectedDir.String(t, fmt.Sprintf("%s_with_%s_error_in_run.out", tt.name, errorInRun), writer.String())
			})
		}
	}
}

func TestStepCreateVirtualMachine_Run_Error(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, testName := range []string{
		"CreateDisk",
		"CreateExternalAddress",
		"CreateNetwork",
		"CreateSubnet",
		"CreateVirtualMachine",
		"CreateFirewallRule",
	} {
		t.Run(testName, func(t *testing.T) {
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

			expectedDiskName := defaultDiskName
			expectedExternalAddressName := defaultExternalAddressName
			expectedNetworkName := defaultNetworkName
			expectedSubnetName := defaultSubnetName
			expectedVirtualMachineName := defaultVirtualMachineName
			expectedFirewallRuleName := mws.FirewallRuleName
			expectedDiskRef := new(computeref.NewDiskRef(testProjectName, expectedDiskName))

			expectedErrors := map[string]error{testName: errors.New("test error")}
			func() {
				driver.EXPECT().CreateDisk(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateDisk"]).Times(1)
				if testName == "CreateDisk" {
					return
				}
				driver.EXPECT().CreateExternalAddress(gomock.Any(), gomock.Any()).
					Return(testExternalAddress, expectedErrors["CreateExternalAddress"]).Times(1)
				if testName == "CreateExternalAddress" {
					return
				}
				driver.EXPECT().CreateNetwork(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateNetwork"]).Times(1)
				if testName == "CreateNetwork" {
					return
				}
				driver.EXPECT().CreateSubnet(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateSubnet"]).Times(1)
				if testName == "CreateSubnet" {
					return
				}
				driver.EXPECT().CreateVirtualMachine(gomock.Any(), gomock.Any()).
					Return(testInternalAddress, expectedErrors["CreateVirtualMachine"]).Times(1)
				if testName == "CreateVirtualMachine" {
					return
				}
				driver.EXPECT().CreateFirewallRule(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateFirewallRule"]).Times(1)
			}()

			step := &mws.StepCreateVirtualMachine{
				GeneratedData: &packerbuilderdata.GeneratedData{State: state},
			}

			requireActionHalt(t, state, step.Run(context.Background(), state))

			func() {
				if testName == "CreateDisk" {
					return
				}
				requireStateGet(t, state, mws.DiskNameKey, expectedDiskName)
				if testName == "CreateExternalAddress" {
					return
				}
				requireStateGet(t, state, mws.ExternalAddressNameKey, expectedExternalAddressName)
				if testName == "CreateNetwork" {
					return
				}
				requireStateGet(t, state, mws.NetworkNameKey, expectedNetworkName)
				if testName == "CreateSubnet" {
					return
				}
				requireStateGet(t, state, mws.SubnetNameKey, expectedSubnetName)
				if testName == "CreateVirtualMachine" {
					return
				}
				requireStateGet(t, state, mws.VirtualMachineNameKey, expectedVirtualMachineName)
				if testName == "CreateFirewallRule" {
					return
				}
				requireStateGet(t, state, mws.FirewallRuleNameKey, expectedFirewallRuleName)

				requireStateGet(t, state, mws.DiskRefKey, expectedDiskRef)
				requireStateGet(t, state, mws.InstanceIpKey, testExternalAddress)
				requireStateGet(t, state, mws.InstanceIdKey, expectedVirtualMachineName)
			}()

			expectedDir.String(t, testName+".out", writer.String())
		})
	}
}

func TestStepCreateVirtualMachine_Cleanup_Error(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, testName := range []string{
		"DeleteFirewallRule",
		"DeleteVirtualMachine",
		"DeleteSubnet",
		"DeleteNetwork",
		"DeleteExternalAddress",
		"DeleteDisk",
	} {
		t.Run(testName, func(t *testing.T) {
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

			expectedDiskName := defaultDiskName
			expectedExternalAddressName := defaultExternalAddressName
			expectedNetworkName := defaultNetworkName
			expectedSubnetName := defaultSubnetName
			expectedVirtualMachineName := defaultVirtualMachineName
			expectedFirewallRuleName := mws.FirewallRuleName

			state.Put(mws.DiskNameKey, expectedDiskName)
			state.Put(mws.ExternalAddressNameKey, expectedExternalAddressName)
			state.Put(mws.NetworkNameKey, expectedNetworkName)
			state.Put(mws.SubnetNameKey, expectedSubnetName)
			state.Put(mws.VirtualMachineNameKey, expectedVirtualMachineName)
			state.Put(mws.FirewallRuleNameKey, expectedFirewallRuleName)

			expectedErrors := map[string]error{testName: errors.New("test error")}
			driver.EXPECT().DeleteFirewallRule(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(expectedErrors["DeleteFirewallRule"]).Times(1)
			driver.EXPECT().DeleteVirtualMachine(gomock.Any(), gomock.Any()).
				Return(expectedErrors["DeleteVirtualMachine"]).Times(1)
			driver.EXPECT().DeleteSubnet(gomock.Any(), gomock.Any(), gomock.Any()).
				Return(expectedErrors["DeleteSubnet"]).Times(1)
			driver.EXPECT().DeleteNetwork(gomock.Any(), gomock.Any()).
				Return(expectedErrors["DeleteNetwork"]).Times(1)
			driver.EXPECT().DeleteExternalAddress(gomock.Any(), gomock.Any()).
				Return(expectedErrors["DeleteExternalAddress"]).Times(1)
			driver.EXPECT().DeleteDisk(gomock.Any(), gomock.Any()).
				Return(expectedErrors["DeleteDisk"]).Times(1)

			step := &mws.StepCreateVirtualMachine{
				GeneratedData: &packerbuilderdata.GeneratedData{State: state},
			}

			step.Cleanup(state)
			expectedDir.String(t, testName+".out", writer.String())
		})
	}
}
