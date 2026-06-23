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
				UseExternalAddress:  true,
			},
		},
		{
			name: "network_set",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				UseExternalAddress: true,
			},
		},
		{
			name: "all_default",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
		},
		{
			name: "no_external_address_all_set",
			config: &mws.Config{
				Project:            testProjectName,
				DiskName:           testDiskName,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				VirtualMachineName: testVirtualMachineName,
				SourceImage:        testSourceImage,
				UseExternalAddress: false,
			},
		},
		{
			name: "no_external_address_default",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				UseExternalAddress: false,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			driver := mockmws.NewMockDriver(ctrl)
			writer, state := prepareState(t, tt.config, driver)

			expectedDiskName := cmp.Or(tt.config.DiskName, defaultDiskName)
			expectedExternalAddressName := cmp.Or(tt.config.ExternalAddressName, defaultExternalAddressName)
			expectedNetworkName := cmp.Or(tt.config.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.config.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := cmp.Or(tt.config.VirtualMachineName, defaultVirtualMachineName)
			expectedFirewallRuleName := mws.FirewallRuleName
			expectedDiskRef := new(computeref.NewDiskRef(tt.config.Project, expectedDiskName))
			var expectedExternalAddressRef *vpcref.ExternalAddressRef
			if tt.config.UseExternalAddress {
				expectedExternalAddressRef = new(vpcref.NewExternalAddressRef(tt.config.Project, expectedExternalAddressName))
			}

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

			if tt.config.UseExternalAddress {
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
			}

			driver.EXPECT().
				CreateVirtualMachine(gomock.Any(), mws.CreateVirtualMachineParams{
					VirtualMachineName: expectedVirtualMachineName,
					VMType:             mws.DefaultVMType,
					Zone:               mws.DefaultZone,
					SSHUsername:        mws.DefaultSSHUsername,
					SSHPublicKey:       testSSHPublicKey,
					DiskRefs:           map[string]*computeref.DiskRef{"boot": expectedDiskRef},
					ExternalAddressRef: expectedExternalAddressRef,
					SubnetRef:          new(vpcref.NewSubnetRef(tt.config.Project, expectedNetworkName, expectedSubnetName)),
				}).
				Return(testInternalAddress, nil).
				Times(1)

			if tt.config.UseExternalAddress {
				driver.EXPECT().
					CreateFirewallRule(gomock.Any(), mws.CreateFirewallRuleParams{
						NetworkName:                   expectedNetworkName,
						FirewallRuleName:              mws.FirewallRuleName,
						VirtualMachineInternalAddress: testInternalAddress,
					}).
					Times(1)
			}

			step := &mws.StepCreateVirtualMachine{
				GeneratedData: &packerbuilderdata.GeneratedData{State: state},
			}

			requireActionContinue(t, state, step.Run(context.Background(), state))
			requireStateGets(t, state,
				map[string]any{
					mws.DiskNameKey:           expectedDiskName,
					mws.NetworkNameKey:        expectedNetworkName,
					mws.SubnetNameKey:         expectedSubnetName,
					mws.VirtualMachineNameKey: expectedVirtualMachineName,
					mws.InstanceIDKey:         expectedVirtualMachineName,
					mws.DiskRefKey:            expectedDiskRef,
				})
			requireGeneratedDataGet(t, state, "SourceProject", tt.config.Project)
			requireGeneratedDataGet(t, state, "SourceImageName", testSourceImage)

			if tt.config.UseExternalAddress {
				requireStateGet(t, state, mws.ExternalAddressNameKey, expectedExternalAddressName)
				requireStateGet(t, state, mws.FirewallRuleNameKey, expectedFirewallRuleName)
				requireStateGet(t, state, mws.InstanceIPKey, testExternalAddress)
			} else {
				requireStateGet(t, state, mws.InstanceIPKey, testInternalAddress)
				requireStateNotSet(t, state, mws.ExternalAddressNameKey)
				requireStateNotSet(t, state, mws.FirewallRuleNameKey)
			}

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
				UseExternalAddress:  true,
			},
		},
		{
			name: "network_set",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				UseExternalAddress: true,
			},
		},
		{
			name: "all_default",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
		},
		{
			name: "no_external_address_all_set",
			config: &mws.Config{
				Project:            testProjectName,
				DiskName:           testDiskName,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				VirtualMachineName: testVirtualMachineName,
				SourceImage:        testSourceImage,
				UseExternalAddress: false,
			},
		},
		{
			name: "no_external_address_default",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				UseExternalAddress: false,
			},
		},
	} {
		possibleErrors := []string{
			"CreateDisk",
			"CreateVirtualMachine",
			"None",
		}
		if tt.config.UseExternalAddress {
			possibleErrors = []string{
				"CreateDisk",
				"CreateExternalAddress",
				"CreateNetwork",
				"CreateSubnet",
				"CreateVirtualMachine",
				"CreateFirewallRule",
				"None",
			}
		}
		for _, errorInRun := range possibleErrors {
			t.Run(tt.name, func(t *testing.T) {
				ctrl := gomock.NewController(t)
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

					if tt.config.UseExternalAddress {
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
					}

					if errorInRun == "CreateVirtualMachine" {
						return
					}
					state.Put(mws.VirtualMachineNameKey, expectedVirtualMachineName)
					driver.EXPECT().DeleteVirtualMachine(gomock.Any(), expectedVirtualMachineName).Times(1)

					if tt.config.UseExternalAddress {
						if errorInRun == "CreateFirewallRule" {
							return
						}
						state.Put(mws.FirewallRuleNameKey, expectedFirewallRuleName)
						driver.EXPECT().DeleteFirewallRule(gomock.Any(), expectedNetworkName, expectedFirewallRuleName).Times(1)
					}
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

	for _, tt := range []struct {
		name      string
		config    *mws.Config
		errorStep string
	}{
		{
			name: "CreateDisk_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "CreateDisk",
		},
		{
			name: "CreateExternalAddress_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "CreateExternalAddress",
		},
		{
			name: "CreateNetwork_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "CreateNetwork",
		},
		{
			name: "CreateSubnet_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "CreateSubnet",
		},
		{
			name: "CreateVirtualMachine_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "CreateVirtualMachine",
		},
		{
			name: "CreateFirewallRule_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "CreateFirewallRule",
		},
		{
			name: "CreateDisk_no_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				UseExternalAddress: false,
			},
			errorStep: "CreateDisk",
		},
		{
			name: "CreateVirtualMachine_no_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				UseExternalAddress: false,
			},
			errorStep: "CreateVirtualMachine",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			driver := mockmws.NewMockDriver(ctrl)

			writer, state := prepareState(t, tt.config, driver)

			expectedDiskName := defaultDiskName
			expectedExternalAddressName := defaultExternalAddressName
			expectedNetworkName := cmp.Or(tt.config.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.config.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := defaultVirtualMachineName
			expectedFirewallRuleName := mws.FirewallRuleName
			expectedDiskRef := new(computeref.NewDiskRef(tt.config.Project, expectedDiskName))

			expectedErrors := map[string]error{tt.errorStep: errors.New("test error")}
			requireStateKV := make(map[string]any)
			func() {
				driver.EXPECT().CreateDisk(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateDisk"]).Times(1)
				if tt.errorStep == "CreateDisk" {
					return
				}
				requireStateKV[mws.DiskNameKey] = expectedDiskName

				if tt.config.UseExternalAddress {
					driver.EXPECT().CreateExternalAddress(gomock.Any(), gomock.Any()).
						Return(testExternalAddress, expectedErrors["CreateExternalAddress"]).Times(1)
					if tt.errorStep == "CreateExternalAddress" {
						return
					}
					requireStateKV[mws.ExternalAddressNameKey] = expectedExternalAddressName

					driver.EXPECT().CreateNetwork(gomock.Any(), gomock.Any()).
						Return(expectedErrors["CreateNetwork"]).Times(1)
					if tt.errorStep == "CreateNetwork" {
						return
					}
					requireStateKV[mws.NetworkNameKey] = expectedNetworkName

					driver.EXPECT().CreateSubnet(gomock.Any(), gomock.Any()).
						Return(expectedErrors["CreateSubnet"]).Times(1)
					if tt.errorStep == "CreateSubnet" {
						return
					}
					requireStateKV[mws.SubnetNameKey] = expectedSubnetName
				}

				driver.EXPECT().CreateVirtualMachine(gomock.Any(), gomock.Any()).
					Return(testInternalAddress, expectedErrors["CreateVirtualMachine"]).Times(1)
				if tt.errorStep == "CreateVirtualMachine" {
					return
				}
				requireStateKV[mws.VirtualMachineNameKey] = expectedVirtualMachineName

				if tt.config.UseExternalAddress {
					driver.EXPECT().CreateFirewallRule(gomock.Any(), gomock.Any()).
						Return(expectedErrors["CreateFirewallRule"]).Times(1)
				}
				if tt.errorStep == "CreateFirewallRule" {
					return
				}
				if tt.config.UseExternalAddress {
					requireStateKV[mws.FirewallRuleNameKey] = expectedFirewallRuleName
					requireStateKV[mws.InstanceIPKey] = testExternalAddress
				} else {
					requireStateKV[mws.InstanceIPKey] = testInternalAddress
				}
				requireStateKV[mws.InstanceIDKey] = expectedVirtualMachineName
				requireStateKV[mws.DiskRefKey] = expectedDiskRef
			}()

			step := &mws.StepCreateVirtualMachine{
				GeneratedData: &packerbuilderdata.GeneratedData{State: state},
			}

			requireActionHalt(t, state, step.Run(context.Background(), state))
			requireStateGets(t, state, requireStateKV)

			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func TestStepCreateVirtualMachine_Cleanup_Error(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name      string
		config    *mws.Config
		errorStep string
	}{
		{
			name: "DeleteFirewallRule_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "DeleteFirewallRule",
		},
		{
			name: "DeleteVirtualMachine_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "DeleteVirtualMachine",
		},
		{
			name: "DeleteSubnet_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "DeleteSubnet",
		},
		{
			name: "DeleteNetwork_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "DeleteNetwork",
		},
		{
			name: "DeleteExternalAddress_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "DeleteExternalAddress",
		},
		{
			name: "DeleteDisk_use_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			errorStep: "DeleteDisk",
		},
		{
			name: "DeleteVirtualMachine_no_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				UseExternalAddress: false,
			},
			errorStep: "DeleteVirtualMachine",
		},
		{
			name: "DeleteDisk_no_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				UseExternalAddress: false,
			},
			errorStep: "DeleteDisk",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			driver := mockmws.NewMockDriver(ctrl)

			writer, state := prepareState(t, tt.config, driver)

			expectedDiskName := defaultDiskName
			expectedExternalAddressName := defaultExternalAddressName
			expectedNetworkName := cmp.Or(tt.config.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.config.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := defaultVirtualMachineName
			expectedFirewallRuleName := mws.FirewallRuleName

			state.Put(mws.DiskNameKey, expectedDiskName)
			state.Put(mws.NetworkNameKey, expectedNetworkName)
			state.Put(mws.SubnetNameKey, expectedSubnetName)
			state.Put(mws.VirtualMachineNameKey, expectedVirtualMachineName)
			if tt.config.UseExternalAddress {
				state.Put(mws.ExternalAddressNameKey, expectedExternalAddressName)
				state.Put(mws.FirewallRuleNameKey, expectedFirewallRuleName)
			}

			expectedErrors := map[string]error{tt.errorStep: errors.New("test error")}

			driver.EXPECT().DeleteDisk(gomock.Any(), expectedDiskName).
				Return(expectedErrors["DeleteDisk"]).Times(1)
			driver.EXPECT().DeleteVirtualMachine(gomock.Any(), expectedVirtualMachineName).
				Return(expectedErrors["DeleteVirtualMachine"]).Times(1)
			if tt.config.UseExternalAddress {
				driver.EXPECT().DeleteExternalAddress(gomock.Any(), expectedExternalAddressName).
					Return(expectedErrors["DeleteExternalAddress"]).Times(1)
				driver.EXPECT().DeleteNetwork(gomock.Any(), expectedNetworkName).
					Return(expectedErrors["DeleteNetwork"]).Times(1)
				driver.EXPECT().DeleteSubnet(gomock.Any(), expectedNetworkName, expectedSubnetName).
					Return(expectedErrors["DeleteSubnet"]).Times(1)
				driver.EXPECT().DeleteFirewallRule(gomock.Any(), expectedNetworkName, expectedFirewallRuleName).
					Return(expectedErrors["DeleteFirewallRule"]).Times(1)
			}

			step := &mws.StepCreateVirtualMachine{
				GeneratedData: &packerbuilderdata.GeneratedData{State: state},
			}

			step.Cleanup(state)
			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}
