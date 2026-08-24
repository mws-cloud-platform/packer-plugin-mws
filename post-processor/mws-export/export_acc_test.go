// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/random"
	"github.com/mws-cloud-platform/packer-plugin-mws/example"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/go-sdk/mws"
	computeclient "go.mws.cloud/go-sdk/service/compute/client"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"
)

func TestAccMWSExport(t *testing.T) {
	if os.Getenv(acctest.TestEnvVar) == "" {
		t.Skipf("Acceptance tests skipped unless env '%s' set", acctest.TestEnvVar)
	}
	accessKey := os.Getenv("PKR_VAR_access_key")
	require.NotZero(t, accessKey, "PKR_VAR_access_key env is required prerequisite")
	secretKey := os.Getenv("PKR_VAR_secret_key")
	require.NotZero(t, secretKey, "PKR_VAR_secret_key env is required prerequisite")
	objectStorageBucket := os.Getenv("PKR_VAR_object_storage_bucket")
	require.NotZero(t, objectStorageBucket, "PKR_VAR_object_storage_bucket env is required prerequisite")

	ctx := t.Context()
	sdk, err := mws.Load(ctx)
	require.NoError(t, err, "load MWS sdk")
	imageClient, err := computesdk.NewImage(ctx, sdk)
	require.NoError(t, err, "create image client")

	imageNameWithBuilder := fmt.Sprintf("packer-acctest-%s-image", random.AlphaNumLower(6))
	imageNameWithoutBuilder := "image-for-export-test"
	_, err = imageClient.GetImage(ctx, computeclient.GetImageRequest{
		Image: imageNameWithoutBuilder,
	})
	require.NoErrorf(t, err, "%q image is required prerequisite", imageNameWithoutBuilder)

	awsClient, err := loadAWSClient(ctx, accessKey, secretKey)
	require.NoError(t, err, "load AWS client")

	testCases := []acctest.PluginTestCase{
		{
			Name:     "export_with_builder_example",
			Template: example.ExportHCL,
			Type:     "mws",
			BuildExtraArgs: []string{
				"-var", "image_name=" + imageNameWithBuilder,
			},
			Check:    check(ctx, awsClient, objectStorageBucket, imageNameWithBuilder),
			Teardown: teardown(ctx, awsClient, objectStorageBucket, imageNameWithBuilder),
		},
		{
			Name:     "export_of_existing_image_example",
			Template: example.ExportOfExistingImageHCL,
			Type:     "mws",
			BuildExtraArgs: []string{
				"-var", "image_name=" + imageNameWithoutBuilder,
			},
			Check:    check(ctx, awsClient, objectStorageBucket, imageNameWithoutBuilder),
			Teardown: teardown(ctx, awsClient, objectStorageBucket, imageNameWithoutBuilder),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()
			acctest.TestPlugin(t, &testCase)
		})
	}
}

func loadAWSClient(ctx context.Context, accessKey, secretKey string) (*s3.Client, error) {
	creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(config.DefaultObjectStorageRegion),
		awsconfig.WithCredentialsProvider(creds),
		awsconfig.WithBaseEndpoint(config.DefaultObjectStorageEndpoint),
	)
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}

	awsClient := s3.NewFromConfig(awsConfig)

	return awsClient, nil
}

func check(ctx context.Context, awsClient *s3.Client, objectStorageBucket, imageName string) func(buildCommand *exec.Cmd, logfile string) error {
	return func(buildCommand *exec.Cmd, logfile string) error {
		if buildCommand.ProcessState != nil && buildCommand.ProcessState.ExitCode() != 0 {
			return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
		}
		if _, err := awsClient.HeadObject(ctx, &s3.HeadObjectInput{
			Bucket: &objectStorageBucket,
			Key:    new(imageName + ".qcow2"),
		}); err != nil {
			return fmt.Errorf("get image from object storage: %w", err)
		}
		return nil
	}
}

func teardown(ctx context.Context, awsClient *s3.Client, objectStorageBucket, imageName string) func() error {
	return func() error {
		if _, err := awsClient.DeleteObject(ctx, &s3.DeleteObjectInput{
			Bucket: &objectStorageBucket,
			Key:    new(imageName + ".qcow2"),
		}); err != nil {
			return fmt.Errorf("delete image from object storage: %w", err)
		}

		return nil
	}
}
