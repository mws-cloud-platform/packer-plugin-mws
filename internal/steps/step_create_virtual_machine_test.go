// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps_test

import (
	"bytes"
	"cmp"
	"errors"
	"io/fs"
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
	testSourceDiskBackup    = "test-source-disk-backup"
	testDiskSize            = "10 GB"
	testImageForExport      = "test-image-for-export"
	testImageForExportSize  = "20 GB"

	defaultExportDiskName      = packerPrefix + "disk-for-export"
	defaultBootDiskName        = packerPrefix + "boot-disk"
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

//nolint:cyclop // Test function with complex logic
func TestStepCreateVirtualMachine_Run(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name                 string
		step                 *steps.StepCreateVirtualMachine
		expectedBootDiskSize *bytesize.ByteSize
		errorStep            string
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
			name: "with_export_disk",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
				DiskForExportConfig: &commonconfig.DiskForExportConfig{
					ImageForExport: testImageForExport,
				},
			},
			expectedBootDiskSize: new(bytesize.MustNewFromInt64(30, bytesize.GB)),
		},
		{
			name: "with_export_disk_and_explicit_disk_size",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						DiskSize: testDiskSize,
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
				DiskForExportConfig: &commonconfig.DiskForExportConfig{
					ImageForExport: testImageForExport,
				},
			},
		},
		{
			name: "with_source_disk_backup_and_export_disk",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					DiskConfig: commonconfig.DiskConfig{
						SourceDiskBackup: "test-source-disk-backup",
					},
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
				DiskForExportConfig: &commonconfig.DiskForExportConfig{
					ImageForExport: testImageForExport,
				},
			},
			expectedBootDiskSize: new(bytesize.MustNewFromInt64(30, bytesize.GB)),
		},
		{
			name: "error_getting_source_image_min_disk_size",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
				DiskForExportConfig: &commonconfig.DiskForExportConfig{
					ImageForExport: testImageForExport,
				},
			},
			errorStep: "GetImageMinDiskSizeSource",
		},
		{
			name: "error_getting_export_image_min_disk_size",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
				DiskForExportConfig: &commonconfig.DiskForExportConfig{
					ImageForExport: testImageForExport,
				},
			},
			errorStep: "GetImageMinDiskSizeExport",
		},
		{
			name: "error_creating_export_disk",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
				DiskForExportConfig: &commonconfig.DiskForExportConfig{
					ImageForExport: testImageForExport,
				},
			},
			errorStep: "CreateExportDisk",
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
			name: "error_at_CreateBootDisk_use_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
			},
			errorStep: "CreateBootDisk",
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
			name: "error_at_CreateBootDisk_no_external_address",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						NetworkName:        testNetworkName,
						SubnetName:         testSubnetName,
						UseExternalAddress: false,
					},
				},
			},
			errorStep: "CreateBootDisk",
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
			t.Parallel()
			ctrl := gomock.NewController(t)
			driver := mocksteps.NewMockStepCreateVirtualMachineDriver(ctrl)
			writer, state := prepareState(driver)
			prepareStep(t, tt.step, state)

			expectedExportDiskName := defaultExportDiskName
			expectedBootDiskName := cmp.Or(tt.step.DiskName, defaultBootDiskName)
			expectedExternalAddressName := cmp.Or(tt.step.ExternalAddressName, defaultExternalAddressName)
			expectedNetworkName := cmp.Or(tt.step.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.step.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := cmp.Or(tt.step.VirtualMachineName, defaultVirtualMachineName)
			expectedFirewallRuleName := defaultFirewallRuleName

			var expectedExportDiskRef *computeref.DiskRef
			expectedBootDiskRef := new(computeref.NewDiskRef(tt.step.Project, expectedBootDiskName))
			expectedInstanceIP := testInternalAddress
			var expectedExternalAddressRef *vpcref.ExternalAddressRef
			if tt.step.UseExternalAddress {
				expectedInstanceIP = testExternalAddress
				expectedExternalAddressRef = new(vpcref.NewExternalAddressRef(tt.step.Project, expectedExternalAddressName))
			}

			expectedErrors := map[string]error{tt.errorStep: errors.New("test error")}
			requireStateKV := make(map[string]any)
			func() {
				var expectedBootDiskSize *bytesize.ByteSize
				if tt.step.DiskSize != "" {
					expectedBootDiskSize = new(bytesize.MustParseString(tt.step.DiskSize))
				} else {
					sourceImageSize := new(bytesize.MustParseString(testDiskSize))
					sourceDiskBackupSize := new(bytesize.MustParseString(testDiskSize))
					exportImageSize := new(bytesize.MustParseString(testImageForExportSize))
					zeroSize := new(bytesize.MustNewFromInt64(0, bytesize.B))

					driver.EXPECT().
						GetImageMinDiskSize(gomock.Any(), nil).
						Return(zeroSize, nil).
						AnyTimes()

					if tt.step.DiskForExportConfig != nil && tt.step.ImageForExport != "" {
						imageForExportProject := cmp.Or(tt.step.ImageForExportProject, tt.step.Project)
						driver.EXPECT().
							GetImageMinDiskSize(gomock.Any(), new(computeref.NewImageRef(imageForExportProject, tt.step.DiskForExportConfig.ImageForExport))).
							Return(exportImageSize, expectedErrors["GetImageMinDiskSizeExport"]).
							MaxTimes(1)
						if tt.errorStep == "GetImageMinDiskSizeExport" {
							return
						}
					}

					if tt.step.SourceImage != "" {
						driver.EXPECT().
							GetImageMinDiskSize(gomock.Any(), new(computeref.NewImageRef(tt.step.Project, tt.step.SourceImage))).
							Return(sourceImageSize, expectedErrors["GetImageMinDiskSizeSource"]).
							MaxTimes(1)
						if tt.errorStep == "GetImageMinDiskSizeSource" {
							return
						}
					}

					if tt.step.SourceDiskBackup != "" {
						driver.EXPECT().
							GetDiskBackupMinDiskSize(gomock.Any(), new(computeref.NewDiskBackupRef(tt.step.Project, tt.step.SourceDiskBackup))).
							Return(sourceDiskBackupSize, expectedErrors["GetDiskBackupMinDiskSizeSource"]).
							MaxTimes(1)
						if tt.errorStep == "GetDiskBackupMinDiskSizeSource" {
							return
						}
					}

					expectedBootDiskSize = new(bytesize.MustNewFromInt64(10, bytesize.GB))
					if tt.expectedBootDiskSize != nil {
						expectedBootDiskSize = tt.expectedBootDiskSize
					}
				}

				if tt.step.DiskForExportConfig != nil {
					expectedExportDiskRef = new(computeref.NewDiskRef(tt.step.Project, expectedExportDiskName))
					createExportDiskParams := drivermws.CreateDiskParams{
						DiskName: expectedExportDiskName,
						DiskType: commonconfig.DefaultDiskType,
						Iops:     commonconfig.DefaultDiskIOPS,
						ImageRef: new(computeref.NewImageRef(tt.step.Project, testImageForExport)),
						Zone:     commonconfig.DefaultZone,
					}
					driver.EXPECT().
						CreateDisk(gomock.Any(), createExportDiskParams).
						Return(expectedErrors["CreateExportDisk"]).MinTimes(1).MaxTimes(2)
					if expectedErrors["CreateExportDisk"] != nil {
						return
					}
				}

				createBootDiskParams := drivermws.CreateDiskParams{
					DiskName: expectedBootDiskName,
					DiskType: commonconfig.DefaultDiskType,
					Size:     expectedBootDiskSize,
					Iops:     commonconfig.DefaultDiskIOPS,
					Zone:     commonconfig.DefaultZone,
				}

				if tt.step.SourceImage != "" {
					createBootDiskParams.ImageRef = new(computeref.NewImageRef(tt.step.Project, testSourceImage))
				} else if tt.step.SourceDiskBackup != "" {
					createBootDiskParams.DiskBackupRef = new(computeref.NewDiskBackupRef(tt.step.Project, testSourceDiskBackup))
				}

				driver.EXPECT().
					CreateDisk(gomock.Any(), createBootDiskParams).
					Return(expectedErrors["CreateBootDisk"])
				if tt.errorStep == "CreateBootDisk" {
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
						BootDiskRef:        expectedBootDiskRef,
						ExportDiskRef:      expectedExportDiskRef,
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
							VirtualMachineInternalAddress: testInternalAddress,
						}).
						Return(expectedErrors["CreateFirewallRule"]).
						Times(1)
				}
				if tt.errorStep == "CreateFirewallRule" {
					return
				}

				requireStateKV[common.InstanceIPKey] = expectedInstanceIP.String()
				requireStateKV[common.InstanceIDKey] = expectedVirtualMachineName
				requireStateKV[common.DiskRefKey] = expectedBootDiskRef
			}()

			if tt.errorStep != "" {
				testutil.RequireActionHalt(t, state, tt.step.Run(t.Context(), state))
			} else {
				testutil.RequireActionContinue(t, state, tt.step.Run(t.Context(), state))
				testutil.RequireStateGets(t, state, requireStateKV)
				testutil.RequireGeneratedDataGet(t, state, "SourceProject", testProjectName)
				if tt.step.SourceImage != "" {
					testutil.RequireGeneratedDataGet(t, state, "SourceImageName", testSourceImage)
				} else {
					testutil.RequireGeneratedDataGet(t, state, "SourceImageName", testSourceDiskBackup)
				}
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
			name: "with_export_disk",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
				},
				DiskForExportConfig: &commonconfig.DiskForExportConfig{
					ImageForExport: testImageForExport,
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
			name: "serial_console",
			step: &steps.StepCreateVirtualMachine{
				VirtualMachineConfig: commonconfig.VirtualMachineConfig{
					NetworkConfig: commonconfig.NetworkConfig{
						UseExternalAddress: true,
					},
					SerialConsoleLogFile: "serial_console.log",
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
			t.Parallel()
			ctrl := gomock.NewController(t)
			driver := mocksteps.NewMockStepCreateVirtualMachineDriver(ctrl)
			fw := mocksteps.NewMockFileWriter(ctrl)
			writer, state := prepareState(driver)
			prepareStep(t, tt.step, state)
			tt.step.FileWriter = fw

			expectedExportDiskName := defaultExportDiskName
			expectedBootDiskName := cmp.Or(tt.step.DiskName, defaultBootDiskName)
			expectedExternalAddressName := cmp.Or(tt.step.ExternalAddressName, defaultExternalAddressName)
			expectedNetworkName := cmp.Or(tt.step.NetworkName, defaultNetworkName)
			expectedSubnetName := cmp.Or(tt.step.SubnetName, defaultSubnetName)
			expectedVirtualMachineName := cmp.Or(tt.step.VirtualMachineName, defaultVirtualMachineName)
			expectedFirewallRuleName := defaultFirewallRuleName

			if tt.step.SerialConsoleLogFile != "" {
				data := []byte("virtual machine logs...\none more line of logs...")
				driver.EXPECT().GetSerialPortOutput(gomock.Any(), expectedVirtualMachineName, 1).
					Return(data, nil).Times(1)
				fw.EXPECT().WriteFile("serial_console.log", data, fs.FileMode(0644)).
					Return(nil).Times(1)
			}
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
			driver.EXPECT().DeleteDisk(gomock.Any(), expectedBootDiskName).
				Return(tt.expectedErr).Times(1)

			if tt.step.DiskForExportConfig != nil {
				driver.EXPECT().DeleteDisk(gomock.Any(), expectedExportDiskName).
					Return(tt.expectedErr).Times(1)
			}

			tt.step.Cleanup(state)
			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}

func prepareStep(t *testing.T, step *steps.StepCreateVirtualMachine, state multistep.StateBag) {
	step.Project = testProjectName
	step.SourceProject = testProjectName
	if step.SourceDiskBackup == "" {
		step.SourceImage = testSourceImage
	}
	step.Zone = commonconfig.DefaultZone
	step.Communicator = &communicator.Config{
		SSH: communicator.SSH{
			SSHUsername:  commonconfig.DefaultSSHUsername,
			SSHPublicKey: []byte(testSSHPublicKey),
		},
	}

	if step.DiskForExportConfig != nil {
		step.ImageForExportProject = testProjectName
		step.DiskForExportConfig.SetDefaults()
		require.NoError(t, step.DiskForExportConfig.Validate())
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
