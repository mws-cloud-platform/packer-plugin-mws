// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps_test

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/stretchr/testify/require"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	mocksteps "github.com/mws-cloud-platform/packer-plugin-mws/internal/steps/mock"

	commonconfig "github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/steps"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/testutil"
	"go.mws.cloud/go-sdk/pkg/apimodels/cidraddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/ipaddress"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	vpcref "go.mws.cloud/go-sdk/service/resources/references/vpc"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
	"go.uber.org/mock/gomock"
)

const (
	packerPrefix            = "packer-"
	testProjectName         = "test-project"
	testDiskName            = "test-disk"
	testExternalAddressName = "test-external-address"
	testNetworkName         = "test-network"
	testSubnetName          = "test-subnet"
	testVirtualMachineName  = "test-vm"
	testSSHPublicKey        = "test-public-key"
	testSourceImage         = "test-source-image"

	defaultDiskName            = packerPrefix + "disk"
	defaultExternalAddressName = packerPrefix + "external-address"
	defaultNetworkName         = packerPrefix + "network"
	defaultSubnetName          = packerPrefix + "subnet"
	defaultVirtualMachineName  = packerPrefix + "vm"

	errInternal = consterr.Error("internal error")
)

var (
	testInternalAddress = new(ipaddress.MustParseIPAddressString("192.168.0.10"))
	testExternalAddress = new(ipaddress.MustParseIPAddressString("10.20.30.40"))
)

func TestStepCreateVirtualMachine_Run_Success(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())
	for _, tt := range []struct {
		name string
		step *steps.StepCreateVirtualMachine
	}{
		{
			name: "all_set",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						DiskName:    testDiskName,
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:         testNetworkName,
						SubnetName:          testSubnetName,
						ExternalAddressName: testExternalAddressName,
						UseExternalAddress:  true,
					},
					VirtualMachineName: testVirtualMachineName,
				},
			},
		},
		{
			name: "network_set",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						UseExternalAddress: true,
					},
				},
			},
		},
		{
			name: "all_default",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
		},
		{
			name: "no_external_address_all_set",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						DiskName:    testDiskName,
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
					VirtualMachineName: testVirtualMachineName,
				},
			},
		},
		{
			name: "no_external_address_default",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			driver := mocksteps.NewMockStepCreateVirtualMachineDriver(ctrl)
			writer, state := prepareState(driver)
			prepareStep(t, tt.step, state)

			expectedDiskName := cmp.Or(tt.step.DiskName, defaultDiskName)
			expectedExternalAddressName := cmp.Or(tt.step.ExternalAddressName, defaultExternalAddressName)
			expectedNetworkName := cmp.Or(tt.step.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.step.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := cmp.Or(tt.step.VirtualMachineName, defaultVirtualMachineName)
			expectedFirewallRuleName := steps.FirewallRuleName
			expectedDiskRef := new(computeref.NewDiskRef(tt.step.Project, expectedDiskName))
			var expectedExternalAddressRef *vpcref.ExternalAddressRef
			if tt.step.UseExternalAddress {
				expectedExternalAddressRef = new(vpcref.NewExternalAddressRef(tt.step.Project, expectedExternalAddressName))
			}

			driver.EXPECT().
				CreateDisk(gomock.Any(), drivermws.CreateDiskParams{
					DiskName: expectedDiskName,
					DiskType: commonconfig.DefaultDiskType,
					Size:     bytesize.MustParseString(commonconfig.DefaultDiskSize),
					Iops:     commonconfig.DefaultDiskIOPS,
					ImageRef: new(computeref.NewImageRef(tt.step.Project, testSourceImage)),
					Zone:     commonconfig.DefaultZone,
				}).
				Times(1)

			if tt.step.UseExternalAddress {
				driver.EXPECT().
					CreateExternalAddress(gomock.Any(), drivermws.CreateExternalAddressParams{
						ExternalAddressName: expectedExternalAddressName,
					}).
					Return(testExternalAddress, nil).
					Times(1)

				if tt.step.NetworkName == "" {
					driver.EXPECT().
						CreateNetwork(gomock.Any(), drivermws.CreateNetworkParams{
							NetworkName: expectedNetworkName,
						}).
						Times(1)
				}

				if tt.step.SubnetName == "" {
					driver.EXPECT().
						CreateSubnet(gomock.Any(), drivermws.CreateSubnetParams{
							NetworkName: expectedNetworkName,
							SubnetName:  expectedSubnetName,
							SubnetCidr:  cidraddress.MustParseCIDR4AddressString(commonconfig.DefaultSubnetCidr),
						}).
						Times(1)
				}
			}

			driver.EXPECT().
				CreateVirtualMachine(gomock.Any(), drivermws.CreateVirtualMachineParams{
					VirtualMachineName: expectedVirtualMachineName,
					VMType:             commonconfig.DefaultVMType,
					Zone:               commonconfig.DefaultZone,
					SSHUsername:        commonconfig.DefaultSSHUsername,
					SSHPublicKey:       testSSHPublicKey,
					DiskRef:            expectedDiskRef,
					ExternalAddressRef: expectedExternalAddressRef,
					SubnetRef:          new(vpcref.NewSubnetRef(tt.step.Project, expectedNetworkName, expectedSubnetName)),
				}).
				Return(testInternalAddress, nil).
				Times(1)

			if tt.step.UseExternalAddress {
				driver.EXPECT().
					CreateFirewallRule(gomock.Any(), drivermws.CreateFirewallRuleParams{
						NetworkName:                   expectedNetworkName,
						FirewallRuleName:              steps.FirewallRuleName,
						VirtualMachineInternalAddress: testInternalAddress.String(),
					}).
					Times(1)
			}

			testutil.RequireActionContinue(t, state, tt.step.Run(t.Context(), state))
			testutil.RequireStateGets(t, state,
				map[string]any{
					common.DiskNameKey:           expectedDiskName,
					common.NetworkNameKey:        expectedNetworkName,
					common.SubnetNameKey:         expectedSubnetName,
					common.VirtualMachineNameKey: expectedVirtualMachineName,
					common.InstanceIDKey:         expectedVirtualMachineName,
					common.DiskRefKey:            expectedDiskRef,
				})
			testutil.RequireGeneratedDataGet(t, state, "SourceProject", testProjectName)
			testutil.RequireGeneratedDataGet(t, state, "SourceImageName", testSourceImage)

			if tt.step.UseExternalAddress {
				testutil.RequireStateGet(t, state, common.ExternalAddressNameKey, expectedExternalAddressName)
				testutil.RequireStateGet(t, state, common.FirewallRuleNameKey, expectedFirewallRuleName)
				testutil.RequireStateGet(t, state, common.InstanceIPKey, testExternalAddress.String())
			} else {
				testutil.RequireStateGet(t, state, common.InstanceIPKey, testInternalAddress.String())
				testutil.RequireStateNotSet(t, state, common.ExternalAddressNameKey)
				testutil.RequireStateNotSet(t, state, common.FirewallRuleNameKey)
			}

			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func TestStepCreateVirtualMachine_Cleanup_Success(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name string
		step *steps.StepCreateVirtualMachine
	}{
		{
			name: "all_set",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						DiskName: testDiskName,

						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:         testNetworkName,
						SubnetName:          testSubnetName,
						ExternalAddressName: testExternalAddressName,
						UseExternalAddress:  true,
					},
					VirtualMachineName: testVirtualMachineName,
				},
			},
		},
		{
			name: "network_set",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{

					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						UseExternalAddress: true,
					},
				},
			},
		},
		{
			name: "all_default",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
		},
		{
			name: "no_external_address_all_set",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						DiskName:    testDiskName,
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
					VirtualMachineName: testVirtualMachineName,
				},
			},
		},
		{
			name: "no_external_address_default",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
		},
	} {
		possibleErrors := []string{
			"CreateDisk",
			"CreateVirtualMachine",
			"None",
		}
		if tt.step.UseExternalAddress {
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
				driver := mocksteps.NewMockStepCreateVirtualMachineDriver(ctrl)
				writer, state := prepareState(driver)
				prepareStep(t, tt.step, state)

				expectedDiskName := cmp.Or(tt.step.DiskName, defaultDiskName)
				expectedExternalAddressName := cmp.Or(tt.step.ExternalAddressName, defaultExternalAddressName)
				expectedNetworkName := cmp.Or(tt.step.NetworkName, defaultNetworkName)
				expectedSubnetName := cmp.Or(tt.step.SubnetName, defaultSubnetName)
				expectedVirtualMachineName := cmp.Or(tt.step.VirtualMachineName, defaultVirtualMachineName)
				expectedFirewallRuleName := steps.FirewallRuleName

				func() {
					if errorInRun == "CreateDisk" {
						return
					}
					state.Put(common.DiskNameKey, expectedDiskName)
					driver.EXPECT().DeleteDisk(gomock.Any(), expectedDiskName).Times(1)

					if tt.step.UseExternalAddress {
						if errorInRun == "CreateExternalAddress" {
							return
						}
						state.Put(common.ExternalAddressNameKey, expectedExternalAddressName)
						driver.EXPECT().DeleteExternalAddress(gomock.Any(), expectedExternalAddressName).Times(1)
						if errorInRun == "CreateNetwork" {
							return
						}
						state.Put(common.NetworkNameKey, expectedNetworkName)
						if tt.step.NetworkName == "" {
							driver.EXPECT().DeleteNetwork(gomock.Any(), expectedNetworkName).Times(1)
						}
						if errorInRun == "CreateSubnet" {
							return
						}
						state.Put(common.SubnetNameKey, expectedSubnetName)
						if tt.step.SubnetName == "" {
							driver.EXPECT().DeleteSubnet(gomock.Any(), expectedNetworkName, expectedSubnetName).Times(1)
						}
					}

					if errorInRun == "CreateVirtualMachine" {
						return
					}
					state.Put(common.VirtualMachineNameKey, expectedVirtualMachineName)
					driver.EXPECT().DeleteVirtualMachine(gomock.Any(), expectedVirtualMachineName).Times(1)

					if tt.step.UseExternalAddress {
						if errorInRun == "CreateFirewallRule" {
							return
						}
						state.Put(common.FirewallRuleNameKey, expectedFirewallRuleName)
						driver.EXPECT().DeleteFirewallRule(gomock.Any(), expectedNetworkName, expectedFirewallRuleName).Times(1)
					}
				}()

				tt.step.Cleanup(state)
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
		step      *steps.StepCreateVirtualMachine
		errorStep string
	}{
		{
			name: "CreateDisk_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateDisk",
		},
		{
			name: "CreateExternalAddress_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateExternalAddress",
		},
		{
			name: "CreateNetwork_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateNetwork",
		},
		{
			name: "CreateSubnet_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateSubnet",
		},
		{
			name: "CreateVirtualMachine_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateVirtualMachine",
		},
		{
			name: "CreateFirewallRule_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateFirewallRule",
		},
		{
			name: "CreateDisk_no_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
			errorStep: "CreateDisk",
		},
		{
			name: "CreateVirtualMachine_no_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
			errorStep: "CreateVirtualMachine",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			driver := mocksteps.NewMockStepCreateVirtualMachineDriver(ctrl)
			writer, state := prepareState(driver)
			prepareStep(t, tt.step, state)

			expectedDiskName := defaultDiskName
			expectedExternalAddressName := defaultExternalAddressName
			expectedNetworkName := cmp.Or(tt.step.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.step.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := defaultVirtualMachineName
			expectedFirewallRuleName := steps.FirewallRuleName
			expectedDiskRef := new(computeref.NewDiskRef(tt.step.Project, expectedDiskName))

			expectedErrors := map[string]error{tt.errorStep: errors.New("test error")}
			requireStateKV := make(map[string]any)
			func() {
				driver.EXPECT().CreateDisk(gomock.Any(), gomock.Any()).
					Return(expectedErrors["CreateDisk"]).Times(1)
				if tt.errorStep == "CreateDisk" {
					return
				}
				requireStateKV[common.DiskNameKey] = expectedDiskName

				if tt.step.UseExternalAddress {
					driver.EXPECT().CreateExternalAddress(gomock.Any(), gomock.Any()).
						Return(testExternalAddress, expectedErrors["CreateExternalAddress"]).Times(1)
					if tt.errorStep == "CreateExternalAddress" {
						return
					}
					requireStateKV[common.ExternalAddressNameKey] = expectedExternalAddressName

					driver.EXPECT().CreateNetwork(gomock.Any(), gomock.Any()).
						Return(expectedErrors["CreateNetwork"]).Times(1)
					if tt.errorStep == "CreateNetwork" {
						return
					}
					requireStateKV[common.NetworkNameKey] = expectedNetworkName

					driver.EXPECT().CreateSubnet(gomock.Any(), gomock.Any()).
						Return(expectedErrors["CreateSubnet"]).Times(1)
					if tt.errorStep == "CreateSubnet" {
						return
					}
					requireStateKV[common.SubnetNameKey] = expectedSubnetName
				}

				driver.EXPECT().CreateVirtualMachine(gomock.Any(), gomock.Any()).
					Return(testInternalAddress, expectedErrors["CreateVirtualMachine"]).Times(1)
				if tt.errorStep == "CreateVirtualMachine" {
					return
				}
				requireStateKV[common.VirtualMachineNameKey] = expectedVirtualMachineName

				if tt.step.UseExternalAddress {
					driver.EXPECT().CreateFirewallRule(gomock.Any(), gomock.Any()).
						Return(expectedErrors["CreateFirewallRule"]).Times(1)
				}
				if tt.errorStep == "CreateFirewallRule" {
					return
				}
				if tt.step.UseExternalAddress {
					requireStateKV[common.FirewallRuleNameKey] = expectedFirewallRuleName
					requireStateKV[common.InstanceIPKey] = testExternalAddress
				} else {
					requireStateKV[common.InstanceIPKey] = testInternalAddress
				}
				requireStateKV[common.InstanceIDKey] = expectedVirtualMachineName
				requireStateKV[common.DiskRefKey] = expectedDiskRef
			}()

			testutil.RequireActionHalt(t, state, tt.step.Run(t.Context(), state))
			testutil.RequireStateGets(t, state, requireStateKV)

			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func TestStepCreateVirtualMachine_Cleanup_Error(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name      string
		step      *steps.StepCreateVirtualMachine
		errorStep string
	}{
		{
			name: "DeleteFirewallRule_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "DeleteFirewallRule",
		},
		{
			name: "DeleteVirtualMachine_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "DeleteVirtualMachine",
		},
		{
			name: "DeleteSubnet_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "DeleteSubnet",
		},
		{
			name: "DeleteNetwork_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "DeleteNetwork",
		},
		{
			name: "DeleteExternalAddress_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "DeleteExternalAddress",
		},
		{
			name: "DeleteDisk_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "DeleteDisk",
		},
		{
			name: "DeleteVirtualMachine_no_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
			errorStep: "DeleteVirtualMachine",
		},
		{
			name: "DeleteDisk_no_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceImage: testSourceImage,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
			errorStep: "DeleteDisk",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			driver := mocksteps.NewMockStepCreateVirtualMachineDriver(ctrl)
			writer, state := prepareState(driver)
			prepareStep(t, tt.step, state)

			expectedDiskName := defaultDiskName
			expectedExternalAddressName := defaultExternalAddressName
			expectedNetworkName := cmp.Or(tt.step.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.step.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := defaultVirtualMachineName
			expectedFirewallRuleName := steps.FirewallRuleName

			state.Put(common.DiskNameKey, expectedDiskName)
			state.Put(common.NetworkNameKey, expectedNetworkName)
			state.Put(common.SubnetNameKey, expectedSubnetName)
			state.Put(common.VirtualMachineNameKey, expectedVirtualMachineName)
			if tt.step.UseExternalAddress {
				state.Put(common.ExternalAddressNameKey, expectedExternalAddressName)
				state.Put(common.FirewallRuleNameKey, expectedFirewallRuleName)
			}

			expectedErrors := map[string]error{tt.errorStep: errors.New("test error")}

			driver.EXPECT().DeleteDisk(gomock.Any(), expectedDiskName).
				Return(expectedErrors["DeleteDisk"]).Times(1)
			driver.EXPECT().DeleteVirtualMachine(gomock.Any(), expectedVirtualMachineName).
				Return(expectedErrors["DeleteVirtualMachine"]).Times(1)
			if tt.step.UseExternalAddress {
				driver.EXPECT().DeleteExternalAddress(gomock.Any(), expectedExternalAddressName).
					Return(expectedErrors["DeleteExternalAddress"]).Times(1)
				driver.EXPECT().DeleteNetwork(gomock.Any(), expectedNetworkName).
					Return(expectedErrors["DeleteNetwork"]).Times(1)
				driver.EXPECT().DeleteSubnet(gomock.Any(), expectedNetworkName, expectedSubnetName).
					Return(expectedErrors["DeleteSubnet"]).Times(1)
				driver.EXPECT().DeleteFirewallRule(gomock.Any(), expectedNetworkName, expectedFirewallRuleName).
					Return(expectedErrors["DeleteFirewallRule"]).Times(1)
			}

			tt.step.Cleanup(state)
			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func prepareStep(t *testing.T, step *steps.StepCreateVirtualMachine, state multistep.StateBag) {
	step.SetDefaults()
	require.NoError(t, step.Validate())
	step.Project = testProjectName
	step.SourceProject = testProjectName
	step.Zone = commonconfig.DefaultZone
	step.SSHUsername = commonconfig.DefaultSSHUsername
	step.SSHPublicKey = testSSHPublicKey
	step.GeneratedData = &packerbuilderdata.GeneratedData{State: state}
}

func prepareState(driver steps.StepCreateVirtualMachineDriver) (*bytes.Buffer, multistep.StateBag) {
	state := new(multistep.BasicStateBag)
	state.Put(common.DriverKey, driver)
	state.Put(common.PrefixKey, packerPrefix)
	writer := new(bytes.Buffer)
	ui := &packer.BasicUi{
		Writer: writer,
	}
	state.Put(common.UIKey, ui)

	return writer, state
}
