// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"context"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	mockmws "github.com/mws-cloud-platform/packer-plugin-mws/builder/mws/mock"
	"go.mws.cloud/go-sdk/pkg/optional"
	resmodels "go.mws.cloud/go-sdk/pkg/resources/models"
	commonmodel "go.mws.cloud/go-sdk/service/common/model"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.uber.org/mock/gomock"
)

func TestStepCreateImage(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	expectedTestImage := &computemodel.ImageOptionalResponse{
		Metadata: optional.NewOptionalNil(commonmodel.CommonTypedResourceMetadataOptionalResponse{
			Id:          new(resmodels.NewAnyResourceID(testImageName)),
			Description: optional.NewOptional(testImageDescription),
		}),
	}

	expectedDefaultImage := &computemodel.ImageOptionalResponse{
		Metadata: optional.NewOptionalNil(commonmodel.CommonTypedResourceMetadataOptionalResponse{
			Id:          new(resmodels.NewAnyResourceID(defaultImageName)),
			Description: optional.NewOptional(mws.DefaultImageDescription),
		}),
	}

	expectedDiskRef := new(computeref.NewDiskRef(testProjectName, testDiskName))

	for _, tt := range []struct {
		name              string
		config            *mws.Config
		expectedImageName string
		expectedImage     *computemodel.ImageOptionalResponse
		expectedError     bool
		customPreparation func(multistep.StateBag, *mockmws.MockDriverMockRecorder)
	}{
		{
			name: "success_set_name",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				ImageName:          testImageName,
				ImageDescription:   testImageDescription,
				UseExternalAddress: true,
			},
			expectedImageName: testImageName,
			expectedImage:     expectedTestImage,
			customPreparation: func(state multistep.StateBag, driver *mockmws.MockDriverMockRecorder) {
				state.Put(mws.DiskRefKey, expectedDiskRef)

				driver.CreateImage(gomock.Any(), mws.CreateImageParams{
					ImageName:        testImageName,
					ImageDescription: testImageDescription,
					DiskRef:          expectedDiskRef,
				}).
					Return(expectedTestImage, nil).
					Times(1)
			},
		},
		{
			name: "success_set_name_no_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				ImageName:          testImageName,
				ImageDescription:   testImageDescription,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				UseExternalAddress: false,
			},
			expectedImageName: testImageName,
			expectedImage:     expectedTestImage,
			customPreparation: func(state multistep.StateBag, driver *mockmws.MockDriverMockRecorder) {
				state.Put(mws.DiskRefKey, expectedDiskRef)

				driver.CreateImage(gomock.Any(), mws.CreateImageParams{
					ImageName:        testImageName,
					ImageDescription: testImageDescription,
					DiskRef:          expectedDiskRef,
				}).
					Return(expectedTestImage, nil).
					Times(1)
			},
		},
		{
			name: "success_default_name",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			expectedImageName: defaultImageName,
			expectedImage:     expectedDefaultImage,
			customPreparation: func(state multistep.StateBag, driver *mockmws.MockDriverMockRecorder) {
				state.Put(mws.DiskRefKey, expectedDiskRef)

				driver.CreateImage(gomock.Any(), mws.CreateImageParams{
					ImageName:        defaultImageName,
					ImageDescription: mws.DefaultImageDescription,
					DiskRef:          expectedDiskRef,
				}).
					Return(expectedDefaultImage, nil).
					Times(1)
			},
		},
		{
			name: "success_default_name_no_external_address",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				NetworkName:        testNetworkName,
				SubnetName:         testSubnetName,
				UseExternalAddress: false,
			},
			expectedImageName: defaultImageName,
			expectedImage:     expectedDefaultImage,
			customPreparation: func(state multistep.StateBag, driver *mockmws.MockDriverMockRecorder) {
				state.Put(mws.DiskRefKey, expectedDiskRef)

				driver.CreateImage(gomock.Any(), mws.CreateImageParams{
					ImageName:        defaultImageName,
					ImageDescription: mws.DefaultImageDescription,
					DiskRef:          expectedDiskRef,
				}).
					Return(expectedDefaultImage, nil).
					Times(1)
			},
		},
		{
			name: "error_missing_disk_ref",
			config: &mws.Config{
				Project:            testProjectName,
				SourceImage:        testSourceImage,
				UseExternalAddress: true,
			},
			expectedError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			driver := mockmws.NewMockDriver(ctrl)
			writer, state := prepareState(t, tt.config, driver)
			state.Put(mws.VirtualMachineNameKey, defaultVirtualMachineName)
			if tt.customPreparation != nil {
				tt.customPreparation(state, driver.EXPECT())
			}

			step := &mws.StepCreateImage{
				GeneratedData: &packerbuilderdata.GeneratedData{State: state},
			}

			action := step.Run(context.Background(), state)
			if tt.expectedError {
				requireActionHalt(t, state, action)
			} else {
				requireActionContinue(t, state, action)
				requireStateGet(t, state, mws.ImageKey, tt.expectedImage)
				requireGeneratedDataGet(t, state, "ImageName", tt.expectedImageName)
			}
			expectedDir.String(t, tt.name+".out", writer.String())
		})
	}
}
