// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport_test

import (
	"bytes"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	commonconfig "github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	mwsexport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export"
	"github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export/mock"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/go-sdk/pkg/apimodels/units/bytesize"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.uber.org/mock/gomock"
)

const diskForExportName = prefix + "disk-for-export"

var (
	imageForExportRef = computeref.NewImageRef("mws-ubuntu", "ubuntu-latest")
	diskForExportRef  = computeref.NewDiskRef(project, diskForExportName)
)

func TestStepAttachDisk_Run(t *testing.T) {
	t.Parallel()

	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tc := range []struct {
		name           string
		driver         func(*testing.T, *mock.MockDriver) *mock.MockDriver
		expectedAction multistep.StepAction
	}{
		{
			name: "ok",
			driver: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				minDiskSize := bytesize.MustParseString("10GB")
				driver.EXPECT().GetImage(gomock.Any(), imageForExportRef).
					Return(&computemodel.ImageOptionalResponse{
						Status: &computemodel.ImageStatusResponse{
							MinDiskSize: &minDiskSize,
						},
					}, nil)
				driver.EXPECT().CreateDisk(gomock.Any(), drivermws.CreateDiskParams{
					DiskName: diskForExportName,
					DiskType: commonconfig.DefaultDiskType,
					Size:     minDiskSize,
					Iops:     commonconfig.DefaultDiskIOPS,
					ImageRef: &imageForExportRef,
					Zone:     commonconfig.DefaultZone,
				}).Return(nil)
				driver.EXPECT().
					AttachDiskToVirtualMachine(gomock.Any(), vmName, diskForExportRef).
					Return(nil)
				return driver
			},
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "get_image_error",
			driver: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				driver.EXPECT().GetImage(gomock.Any(), imageForExportRef).
					Return(nil, errInternal)
				return driver
			},
			expectedAction: multistep.ActionHalt,
		},
		{
			name: "image_no_min_disk_size",
			driver: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				driver.EXPECT().GetImage(gomock.Any(), imageForExportRef).
					Return(&computemodel.ImageOptionalResponse{
						Status: &computemodel.ImageStatusResponse{},
					}, nil)
				return driver
			},
			expectedAction: multistep.ActionHalt,
		},
		{
			name: "create_disk_error",
			driver: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				minDiskSize := bytesize.MustParseString("10GB")
				driver.EXPECT().GetImage(gomock.Any(), imageForExportRef).
					Return(&computemodel.ImageOptionalResponse{
						Status: &computemodel.ImageStatusResponse{
							MinDiskSize: &minDiskSize,
						},
					}, nil)
				driver.EXPECT().CreateDisk(gomock.Any(), drivermws.CreateDiskParams{
					DiskName: diskForExportName,
					DiskType: commonconfig.DefaultDiskType,
					Size:     minDiskSize,
					Iops:     commonconfig.DefaultDiskIOPS,
					ImageRef: &imageForExportRef,
					Zone:     commonconfig.DefaultZone,
				}).Return(errInternal)
				return driver
			},
			expectedAction: multistep.ActionHalt,
		},
		{
			name: "attach_disk_error",
			driver: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				minDiskSize := bytesize.MustParseString("10GB")
				driver.EXPECT().GetImage(gomock.Any(), imageForExportRef).
					Return(&computemodel.ImageOptionalResponse{
						Status: &computemodel.ImageStatusResponse{
							MinDiskSize: &minDiskSize,
						},
					}, nil)
				driver.EXPECT().CreateDisk(gomock.Any(), drivermws.CreateDiskParams{
					DiskName: diskForExportName,
					DiskType: commonconfig.DefaultDiskType,
					Size:     minDiskSize,
					Iops:     commonconfig.DefaultDiskIOPS,
					ImageRef: &imageForExportRef,
					Zone:     commonconfig.DefaultZone,
				}).Return(nil)
				driver.EXPECT().
					AttachDiskToVirtualMachine(gomock.Any(), vmName, diskForExportRef).
					Return(errInternal)
				return driver
			},
			expectedAction: multistep.ActionHalt,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			state := new(multistep.BasicStateBag)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(mws.UIKey, ui)
			state.Put(mws.PrefixKey, prefix)
			driver := mock.NewMockDriver(ctrl)
			state.Put(mws.DriverKey, tc.driver(t, driver))
			state.Put(mws.InstanceIDKey, vmName)

			step := &mwsexport.StepAttachDisk{
				Project:        project,
				Zone:           commonconfig.DefaultZone,
				DiskType:       commonconfig.DefaultDiskType,
				DiskIOPS:       commonconfig.DefaultDiskIOPS,
				ImageRef:       imageForExportRef,
				CleanupTimeout: cleanupTimeout,
			}

			action := step.Run(t.Context(), state)
			require.Equal(t, tc.expectedAction, action)
			dir.String(t, tc.name+".out", writer.String())
		})
	}
}

func TestStepAttachDisk_Cleanup(t *testing.T) {
	t.Parallel()

	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tc := range []struct {
		name                string
		diskExists          bool
		configureDriverMock func(*testing.T, *mock.MockDriver) *mock.MockDriver
	}{
		{
			name:       "ok",
			diskExists: true,
			configureDriverMock: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				t.Helper()
				driver.EXPECT().DetachSecondaryDisksFromVirtualMachine(gomock.Any(), vmName).Return(nil)
				driver.EXPECT().DeleteDisk(gomock.Any(), diskForExportName).Return(nil)
				return driver
			},
		},
		{
			name:       "no_disk",
			diskExists: false,
			configureDriverMock: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				t.Helper()
				return driver
			},
		},
		{
			name:       "detach_error",
			diskExists: true,
			configureDriverMock: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				t.Helper()
				driver.EXPECT().DetachSecondaryDisksFromVirtualMachine(gomock.Any(), vmName).
					Return(errInternal)
				driver.EXPECT().DeleteDisk(gomock.Any(), diskForExportName).Return(nil)
				return driver
			},
		},
		{
			name:       "detach_and_delete_error",
			diskExists: true,
			configureDriverMock: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				t.Helper()
				driver.EXPECT().DetachSecondaryDisksFromVirtualMachine(gomock.Any(), vmName).
					Return(errInternal)
				driver.EXPECT().DeleteDisk(gomock.Any(), diskForExportName).Return(errInternal)
				return driver
			},
		},
		{
			name:       "delete_disk_error",
			diskExists: true,
			configureDriverMock: func(t *testing.T, driver *mock.MockDriver) *mock.MockDriver {
				t.Helper()
				driver.EXPECT().DetachSecondaryDisksFromVirtualMachine(gomock.Any(), vmName).Return(nil)
				driver.EXPECT().DeleteDisk(gomock.Any(), diskForExportName).Return(errInternal)
				return driver
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			state := new(multistep.BasicStateBag)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(mws.UIKey, ui)
			state.Put(mws.PrefixKey, prefix)
			driver := mock.NewMockDriver(ctrl)
			state.Put(mws.DriverKey, tc.configureDriverMock(t, driver))
			state.Put(mws.InstanceIDKey, vmName)
			if tc.diskExists {
				state.Put(mwsexport.DiskForExportNameKey, diskForExportName)
			}

			step := &mwsexport.StepAttachDisk{
				Zone:           commonconfig.DefaultZone,
				DiskType:       commonconfig.DefaultDiskType,
				DiskIOPS:       commonconfig.DefaultDiskIOPS,
				ImageRef:       imageForExportRef,
				CleanupTimeout: cleanupTimeout,
			}

			step.Cleanup(state)
			dir.String(t, tc.name+".out", writer.String())
		})
	}
}
