package mws_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/packerbuilderdata"
	"github.com/stretchr/testify/require"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	mock_mws "github.com/mws-cloud-platform/packer-plugin-mws/builder/mws/mock"
	"go.mws.cloud/go-sdk/pkg/optional"
	resmodels "go.mws.cloud/go-sdk/pkg/resources/models"
	commonmodel "go.mws.cloud/go-sdk/service/common/model"
	computemodel "go.mws.cloud/go-sdk/service/compute/model"
	"go.uber.org/mock/gomock"
)

func TestStepCreateImage_Run_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedImage := &computemodel.ImageOptionalResponse{
		Metadata: optional.NewOptionalNil(commonmodel.CommonTypedResourceMetadataOptionalResponse{
			Id:          new(resmodels.NewAnyResourceID(testImageName)),
			Description: optional.NewOptional(testImageDescription),
		}),
	}

	driver := mock_mws.NewMockDriver(ctrl)
	driver.EXPECT().
		CreateImage(gomock.Any(), mws.CreateImageParams{
			ImageName:        testImageName,
			ImageDescription: testImageDescription,
			DiskRef:          testDiskRef,
		}).
		Return(expectedImage, nil).
		Times(1)

	config := &mws.Config{
		Project:          testProjectName,
		SourceImage:      testSourceImage,
		ImageName:        testImageName,
		ImageDescription: testImageDescription,
	}
	config.SetDefaults()
	require.NoError(t, config.Validate())

	state := new(multistep.BasicStateBag)
	state.Put(mws.ConfigKey, config)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UuidPrefixKey, packerPrefix)
	state.Put(mws.DiskRefKey, testDiskRef)

	var writer bytes.Buffer
	ui := &packer.BasicUi{
		Writer: &writer,
	}
	state.Put(mws.UiKey, ui)

	step := &mws.StepCreateImage{
		GeneratedData: &packerbuilderdata.GeneratedData{State: state},
	}

	requireActionContinue(t, state, step.Run(context.Background(), state))

	requireStateGet(t, state, mws.ImageKey, expectedImage)
	requireGeneratedDataGet(t, state, "ImageName", testImageName)
	requireOutput(t, writer.String())
}

func TestStepCreateImage_Run_WithDefaultImageName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	expectedImage := &computemodel.ImageOptionalResponse{
		Metadata: optional.NewOptionalNil(commonmodel.CommonTypedResourceMetadataOptionalResponse{
			Id:          new(resmodels.NewAnyResourceID(defaultImageName)),
			Description: optional.NewOptional(testImageDescription),
		}),
	}

	driver := mock_mws.NewMockDriver(ctrl)
	driver.EXPECT().
		CreateImage(gomock.Any(), mws.CreateImageParams{
			ImageName:        defaultImageName,
			ImageDescription: testImageDescription,
			DiskRef:          testDiskRef,
		}).
		Return(expectedImage, nil).
		Times(1)

	config := &mws.Config{
		Project:          testProjectName,
		SourceImage:      testSourceImage,
		ImageDescription: testImageDescription,
	}
	config.SetDefaults()
	require.NoError(t, config.Validate())

	state := new(multistep.BasicStateBag)
	state.Put(mws.ConfigKey, config)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UuidPrefixKey, packerPrefix)
	state.Put(mws.DiskRefKey, testDiskRef)

	var writer bytes.Buffer
	ui := &packer.BasicUi{
		Writer: &writer,
	}
	state.Put(mws.UiKey, ui)

	step := &mws.StepCreateImage{
		GeneratedData: &packerbuilderdata.GeneratedData{State: state},
	}

	requireActionContinue(t, state, step.Run(context.Background(), state))
	requireStateGet(t, state, mws.ImageKey, expectedImage)
	requireGeneratedDataGet(t, state, "ImageName", defaultImageName)
	requireOutput(t, writer.String())
}

func TestStepCreateImage_Run_missingDiskRef(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	driver := mock_mws.NewMockDriver(ctrl)

	config := &mws.Config{
		Project:     testProjectName,
		SourceImage: testSourceImage,
	}
	config.SetDefaults()
	require.NoError(t, config.Validate())

	state := new(multistep.BasicStateBag)
	state.Put(mws.ConfigKey, config)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UuidPrefixKey, packerPrefix)

	var writer bytes.Buffer
	ui := &packer.BasicUi{
		Writer: &writer,
	}
	state.Put(mws.UiKey, ui)

	step := &mws.StepCreateImage{
		GeneratedData: &packerbuilderdata.GeneratedData{State: state},
	}

	requireActionHalt(t, state, step.Run(context.Background(), state))
	requireOutput(t, writer.String())
}
