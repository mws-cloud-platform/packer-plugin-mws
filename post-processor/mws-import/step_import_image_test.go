// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport

import (
	"bytes"
	"context"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	mws "github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/testutil"
	mockmws "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-import/mock"
	"go.mws.cloud/go-sdk/pkg/optional"
	resmodels "go.mws.cloud/go-sdk/pkg/resources/models"
	commonmodel "go.mws.cloud/go-sdk/service/common/model"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
	"go.uber.org/mock/gomock"
)

const (
	packerPrefix         = "packer-"
	testProjectName      = "test-project"
	testImageName        = "test-image"
	testImageDisplayName = "Test Image Display Name"
	testImageDescription = "Test image description"
	testExternalURL      = "https://storage.test.mwsapis.ru/test-bucket/path/to/image.qcow2"

	errInternal = consterr.Error("internal error")
)

func TestStepImportImage(t *testing.T) {
	t.Parallel()
	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	image := &computemodel.ImageOptionalResponse{
		Metadata: optional.NewOptionalNil(commonmodel.CommonTypedResourceMetadataOptionalResponse{
			Id:          new(resmodels.NewAnyResourceID(testImageName)),
			Description: optional.NewOptional(testImageDescription),
		}),
	}

	for _, tt := range []struct {
		name             string
		project          string
		imageName        string
		imageDisplayName string
		imageDescription string
		prepare          func(multistep.StateBag, *mockmws.MockDriver)
		expectedError    bool
		expectedImage    *computemodel.ImageOptionalResponse
	}{
		{
			name:             "success",
			project:          testProjectName,
			imageName:        testImageName,
			imageDisplayName: testImageDisplayName,
			imageDescription: testImageDescription,
			prepare: func(state multistep.StateBag, driver *mockmws.MockDriver) {
				state.Put(ExternalURLKey, testExternalURL)

				driver.EXPECT().
					ImportImage(gomock.Any(), drivermws.ImportImageParams{
						ImageName:        testImageName,
						ImageDisplayName: testImageDisplayName,
						ImageDescription: testImageDescription,
						ExternalURL:      testExternalURL,
					}).
					Return(image, nil).
					Times(1)
			},
			expectedImage: image,
		},
		{
			name:             "import_image_error",
			project:          testProjectName,
			imageName:        testImageName,
			imageDisplayName: testImageDisplayName,
			imageDescription: testImageDescription,
			prepare: func(state multistep.StateBag, driver *mockmws.MockDriver) {
				state.Put(ExternalURLKey, testExternalURL)

				driver.EXPECT().
					ImportImage(gomock.Any(), drivermws.ImportImageParams{
						ImageName:        testImageName,
						ImageDisplayName: testImageDisplayName,
						ImageDescription: testImageDescription,
						ExternalURL:      testExternalURL,
					}).
					Return(nil, errInternal).
					Times(1)
			},
			expectedError: true,
		},
		{
			name:             "missing_external_url",
			project:          testProjectName,
			imageName:        testImageName,
			imageDisplayName: testImageDisplayName,
			imageDescription: testImageDescription,
			expectedError:    true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			driver := mockmws.NewMockDriver(ctrl)

			state := new(multistep.BasicStateBag)
			state.Put(mws.DriverKey, driver)
			state.Put(mws.PrefixKey, packerPrefix)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(mws.UIKey, ui)

			if tt.prepare != nil {
				tt.prepare(state, driver)
			}

			step := &StepImportImage{
				Project:          tt.project,
				ImageName:        tt.imageName,
				ImageDisplayName: tt.imageDisplayName,
				ImageDescription: tt.imageDescription,
				GeneratedData:    &packerbuilderdata.GeneratedData{State: state},
			}

			action := step.Run(context.Background(), state)
			if tt.expectedError {
				testutil.RequireActionHalt(t, state, action)
			} else {
				testutil.RequireActionContinue(t, state, action)
				testutil.RequireStateGet(t, state, mws.ImageKey, tt.expectedImage)
				testutil.RequireGeneratedDataGet(t, state, "ImageProject", tt.project)
				testutil.RequireGeneratedDataGet(t, state, "ImageName", tt.imageName)
			}
			dir.String(t, tt.name+".out", writer.String())
		})
	}
}
