// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps_test

import (
	"bytes"
	"cmp"
	"errors"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/communicator"
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

	mwserrors "go.mws.cloud/go-sdk/mws/errors"
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
	testDiskSize            = "10 GB"

	defaultDiskName            = packerPrefix + "boot-disk"
	defaultExternalAddressName = packerPrefix + "external-address"
	defaultNetworkName         = packerPrefix + "network"
	defaultSubnetName          = packerPrefix + "subnet"
	defaultVirtualMachineName  = packerPrefix + "vm"
	defaultFirewallRuleName    = packerPrefix + "ssh-access"

	errInternal = consterr.Error("internal error")
)

var (
	testInternalAddress = new(ipaddress.MustParseIPAddressString("192.168.0.10"))
	testExternalAddress = new(ipaddress.MustParseIPAddressString("10.20.30.40"))
)

func TestStepCreateVirtualMachine_Run(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name      string
		step      *steps.StepCreateVirtualMachine
		errorStep string
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
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
		},
		{
			name: "error_at_CreateDisk_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateDisk",
		},
		{
			name: "error_at_CreateExternalAddress_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateExternalAddress",
		},
		{
			name: "error_at_CreateNetwork_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateNetwork",
		},
		{
			name: "error_at_CreateSubnet_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateSubnet",
		},
		{
			name: "error_at_CreateVirtualMachine_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateVirtualMachine",
		},
		{
			name: "error_at_CreateFirewallRule_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateFirewallRule",
		},
		{
			name: "error_at_CreateDisk_no_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
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
			name: "error_at_CreateVirtualMachine_no_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
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

			expectedDiskName := cmp.Or(tt.step.DiskName, defaultDiskName)
			expectedExternalAddressName := cmp.Or(tt.step.ExternalAddressName, defaultExternalAddressName)
			expectedNetworkName := cmp.Or(tt.step.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.step.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := cmp.Or(tt.step.VirtualMachineName, defaultVirtualMachineName)
			expectedFirewallRuleName := defaultFirewallRuleName

			expectedDiskRef := new(computeref.NewDiskRef(tt.step.Project, expectedDiskName))
			expectedInstanceIP := testInternalAddress
			var expectedExternalAddressRef *vpcref.ExternalAddressRef
			if tt.step.UseExternalAddress {
				expectedInstanceIP = testExternalAddress
				expectedExternalAddressRef = new(vpcref.NewExternalAddressRef(tt.step.Project, expectedExternalAddressName))
			}

			expectedErrors := map[string]error{tt.errorStep: errors.New("test error")}
			requireStateKV := make(map[string]any)
			func() {
				driver.EXPECT().
					CreateDisk(gomock.Any(), drivermws.CreateDiskParams{
						DiskName: expectedDiskName,
						DiskType: commonconfig.DefaultDiskType,
						Size:     new(bytesize.MustParseString(testDiskSize)),
						Iops:     commonconfig.DefaultDiskIOPS,
						ImageRef: new(computeref.NewImageRef(tt.step.Project, testSourceImage)),
						Zone:     commonconfig.DefaultZone,
					}).
					Return(expectedErrors["CreateDisk"]).
					Times(1)
				if tt.errorStep == "CreateDisk" {
					return
				}

				if tt.step.UseExternalAddress {
					driver.EXPECT().
						CreateExternalAddress(gomock.Any(), drivermws.CreateExternalAddressParams{
							ExternalAddressName: expectedExternalAddressName,
						}).
						Return(testExternalAddress, expectedErrors["CreateExternalAddress"]).
						Times(1)
					if tt.errorStep == "CreateExternalAddress" {
						return
					}
				}

				if tt.step.NetworkName == "" {
					driver.EXPECT().
						CreateNetwork(gomock.Any(), drivermws.CreateNetworkParams{
							NetworkName: expectedNetworkName,
						}).
						Return(expectedErrors["CreateNetwork"]).
						Times(1)
					if tt.errorStep == "CreateNetwork" {
						return
					}
				}

				if tt.step.SubnetName == "" {
					driver.EXPECT().
						CreateSubnet(gomock.Any(), drivermws.CreateSubnetParams{
							NetworkName: expectedNetworkName,
							SubnetName:  expectedSubnetName,
							SubnetCidr:  cidraddress.MustParseCIDR4AddressString(commonconfig.DefaultSubnetCidr),
						}).
						Return(expectedErrors["CreateSubnet"]).
						Times(1)
					if tt.errorStep == "CreateSubnet" {
						return
					}
				}

				driver.EXPECT().
					CreateVirtualMachine(gomock.Any(), drivermws.CreateVirtualMachineParams{
						VirtualMachineName: expectedVirtualMachineName,
						VMType:             commonconfig.DefaultVMType,
						Zone:               commonconfig.DefaultZone,
						SSHUsername:        commonconfig.DefaultSSHUsername,
						SSHPublicKey:       testSSHPublicKey,
						BootDiskRef:        expectedDiskRef,
						ExternalAddressRef: expectedExternalAddressRef,
						SubnetRef:          new(vpcref.NewSubnetRef(tt.step.Project, expectedNetworkName, expectedSubnetName)),
					}).
					Return(testInternalAddress, expectedErrors["CreateVirtualMachine"]).
					Times(1)
				if tt.errorStep == "CreateVirtualMachine" {
					return
				}

				if tt.step.UseExternalAddress {
					driver.EXPECT().
						CreateFirewallRule(gomock.Any(), drivermws.CreateFirewallRuleParams{
							NetworkName:                   expectedNetworkName,
							FirewallRuleName:              expectedFirewallRuleName,
							VirtualMachineInternalAddress: testInternalAddress.String(),
						}).
						Return(expectedErrors["CreateFirewallRule"]).
						Times(1)
				}
				if tt.errorStep == "CreateFirewallRule" {
					return
				}

				requireStateKV[common.InstanceIPKey] = expectedInstanceIP.String()
				requireStateKV[common.InstanceIDKey] = expectedVirtualMachineName
				requireStateKV[common.DiskRefKey] = expectedDiskRef
			}()

			if tt.errorStep != "" {
				testutil.RequireActionHalt(t, state, tt.step.Run(t.Context(), state))
			} else {
				testutil.RequireActionContinue(t, state, tt.step.Run(t.Context(), state))
				testutil.RequireStateGets(t, state, requireStateKV)
				testutil.RequireGeneratedDataGet(t, state, "SourceProject", testProjectName)
				testutil.RequireGeneratedDataGet(t, state, "SourceImageName", testSourceImage)
			}

			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func TestStepCreateVirtualMachine_Cleanup(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name        string
		step        *steps.StepCreateVirtualMachine
		expectedErr error
	}{
		{
			name: "full_custom_names",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						DiskName: testDiskName,
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
			name: "use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
		},
		{
			name: "set_network",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						UseExternalAddress: true,
					},
				},
			},
		},
		{
			name: "set_subnet",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: true,
					},
				},
			},
		},
		{
			name: "no_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
		},
		{
			name: "error_not_found",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			expectedErr: mwserrors.NewAPIError(mwserrors.NotFound.HTTPCodes()[0], mwserrors.NotFound, ""),
		},
		{
			name: "error_internal",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			expectedErr: errInternal,
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
			expectedFirewallRuleName := defaultFirewallRuleName

			if tt.step.UseExternalAddress {
				driver.EXPECT().DeleteExternalAddress(gomock.Any(), expectedExternalAddressName).
					Return(tt.expectedErr).Times(1)
			}
			driver.EXPECT().DeleteVirtualMachine(gomock.Any(), expectedVirtualMachineName).
				Return(tt.expectedErr).Times(1)
			if tt.step.SubnetName == "" {
				driver.EXPECT().DeleteSubnet(gomock.Any(), expectedNetworkName, expectedSubnetName).
					Return(tt.expectedErr).Times(1)
			}
			if tt.step.NetworkName == "" {
				driver.EXPECT().DeleteNetwork(gomock.Any(), expectedNetworkName).
					Return(tt.expectedErr).Times(1)
			}
			if tt.step.UseExternalAddress {
				driver.EXPECT().DeleteFirewallRule(gomock.Any(), expectedNetworkName, expectedFirewallRuleName).
					Return(tt.expectedErr).Times(1)
			}
			driver.EXPECT().DeleteDisk(gomock.Any(), expectedDiskName).
				Return(tt.expectedErr).Times(1)

			tt.step.Cleanup(state)
			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func prepareStep(t *testing.T, step *steps.StepCreateVirtualMachine, state multistep.StateBag) {
	step.Project = testProjectName
	step.SourceProject = testProjectName
	step.SourceImage = testSourceImage
	step.DiskSize = testDiskSize
	step.Zone = commonconfig.DefaultZone
	step.Communicator = &communicator.Config{
		SSH: communicator.SSH{
			SSHUsername:  commonconfig.DefaultSSHUsername,
			SSHPublicKey: []byte(testSSHPublicKey),
		},
	}
	step.VirtualMachineConfig.SetDefaults()
	require.NoError(t, step.VirtualMachineConfig.Validate())
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
