// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"bytes"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	mockmws "github.com/mws-cloud-platform/packer-plugin-mws/builder/mws/mock"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/testutil"
	"go.mws.cloud/go-sdk/pkg/optional"
	resmodels "go.mws.cloud/go-sdk/pkg/resources/models"
	commonmodel "go.mws.cloud/go-sdk/service/common/model"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
	"go.uber.org/mock/gomock"
)

const (
	packerPrefix    = "packer-"
	testProjectName = "test-project"
	testDiskName    = "test-disk"

	testImageName        = "test-image"
	testImageDescription = "Test image description"

	defaultVirtualMachineName = packerPrefix + "vm"

	errInternal = consterr.Error("internal error")
)

func TestStepCreateImage(t *testing.T) {
	t.Parallel()
	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	image := &computemodel.ImageOptionalResponse{
		Metadata: optional.NewOptionalNil(commonmodel.CommonTypedResourceMetadataOptionalResponse{
			Id:          new(resmodels.NewAnyResourceID(testImageName)),
			Description: optional.NewOptional(testImageDescription),
		}),
	}

	diskRef := new(computeref.NewDiskRef(testProjectName, testDiskName))

	for _, tt := range []struct {
		name             string
		project          string
		imageName        string
		imageDescription string
		prepare          func(multistep.StateBag, *mockmws.MockDriver)
		expectedError    bool
		expectedImage    *computemodel.ImageOptionalResponse
	}{
		{
			name:             "success",
			project:          testProjectName,
			imageName:        testImageName,
			imageDescription: testImageDescription,
			prepare: func(state multistep.StateBag, driver *mockmws.MockDriver) {
				state.Put(common.DiskRefKey, diskRef)

				driver.EXPECT().
					CreateImage(gomock.Any(), drivermws.CreateImageParams{
						ImageName:        testImageName,
						ImageDescription: testImageDescription,
						DiskRef:          diskRef,
					}).
					Return(image, nil).
					Times(1)
			},
			expectedImage: image,
		},
		{
			name:             "create_image_error",
			project:          testProjectName,
			imageName:        testImageName,
			imageDescription: testImageDescription,
			prepare: func(state multistep.StateBag, driver *mockmws.MockDriver) {
				state.Put(common.DiskRefKey, diskRef)

				driver.EXPECT().
					CreateImage(gomock.Any(), drivermws.CreateImageParams{
						ImageName:        testImageName,
						ImageDescription: testImageDescription,
						DiskRef:          diskRef,
					}).
					Return(nil, errInternal).
					Times(1)
			},
			expectedError: true,
		},
		{
			name:             "missing_disk_ref",
			project:          testProjectName,
			imageName:        testImageName,
			imageDescription: testImageDescription,
			expectedError:    true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			driver := mockmws.NewMockDriver(ctrl)

			state := new(multistep.BasicStateBag)
			state.Put(common.DriverKey, driver)
			state.Put(common.PrefixKey, packerPrefix)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(common.UIKey, ui)
			state.Put(common.VirtualMachineNameKey, defaultVirtualMachineName)

			if tt.prepare != nil {
				tt.prepare(state, driver)
			}

			step := &mws.StepCreateImage{
				Project:          tt.project,
				ImageName:        tt.imageName,
				ImageDescription: tt.imageDescription,
				GeneratedData:    &packerbuilderdata.GeneratedData{State: state},
			}

			action := step.Run(t.Context(), state)
			if tt.expectedError {
				testutil.RequireActionHalt(t, state, action)
			} else {
				testutil.RequireActionContinue(t, state, action)
				testutil.RequireStateGet(t, state, common.ImageKey, tt.expectedImage)
				testutil.RequireGeneratedDataGet(t, state, "ImageProject", tt.project)
				testutil.RequireGeneratedDataGet(t, state, "ImageName", tt.imageName)
			}
			dir.String(t, tt.name+".out", writer.String())
		})
	}
}
