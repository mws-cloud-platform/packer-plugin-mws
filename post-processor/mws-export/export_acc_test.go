// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/random"
	"github.com/mws-cloud-platform/packer-plugin-mws/example"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/go-sdk/mws"
	iamclient "go.mws.cloud/go-sdk/service/iam/client"
	iammodel "go.mws.cloud/go-sdk/service/iam/model"
	iamsdk "go.mws.cloud/go-sdk/service/iam/sdk"
)

func TestAccMWSExport(t *testing.T) {
	ctx := t.Context()
	serviceAccount := os.Getenv("PKR_VAR_service_account")
	objectStorageBucket := os.Getenv("PKR_VAR_object_storage_bucket")

	awsClient, awsCleanup, err := loadAWSClient(ctx, serviceAccount)
	require.NoError(t, err, "load AWS config")
	t.Cleanup(awsCleanup)

	imageName := fmt.Sprintf("packer-acctest-%s-image", random.AlphaNumLower(6))

	testCase := &acctest.PluginTestCase{
		Name:     "export_example",
		Template: example.ExportHCL,
		Type:     "mws",
		BuildExtraArgs: []string{
			"-var", "image_name=" + imageName,
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil && buildCommand.ProcessState.ExitCode() != 0 {
				return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
			}
			if _, err := awsClient.GetObject(ctx, &s3.GetObjectInput{
				Bucket: &objectStorageBucket,
				Key:    new(imageName + ".qcow2"),
			}); err != nil {
				return fmt.Errorf("get image from object storage: %w", err)
			}
			return nil
		},
		Teardown: func() error {
			if _, err := awsClient.DeleteObject(ctx, &s3.DeleteObjectInput{
				Bucket: &objectStorageBucket,
				Key:    new(imageName + ".qcow2"),
			}); err != nil {
				return fmt.Errorf("delete image from object storage: %w", err)
			}

			return nil
		},
	}

	acctest.TestPlugin(t, testCase)

}

func loadAWSClient(ctx context.Context, serviceAccount string) (*s3.Client, func(), error) {
	sdk, err := mws.Load(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("load mws sdk: %w", err)
	}

	hmacKeys, err := iamsdk.NewServiceAccountHmacKey(ctx, sdk)
	if err != nil {
		return nil, nil, fmt.Errorf("create hmac key client: %w", err)
	}

	hmacKeyName := fmt.Sprintf("packer-acctest-%s-hmac-key", random.AlphaNumLower(6))

	hmacKeyResp, err := hmacKeys.CreateHmacKey(ctx, iamclient.UpsertHmacKeyRequest{
		ServiceAccount: serviceAccount,
		KeyName:        hmacKeyName,
		Body: iammodel.HmacKeyRequest{
			Spec: iammodel.HmacKeySpecRequest{
				ExpirationTime: new(time.Now().Add(time.Hour)),
			},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("create hmac key: %w", err)
	}
	hmacAccessKey := hmacKeyResp.GetStatus().GetAccessKeyIdOr("")
	hmacSecretKey := hmacKeyResp.GetStatus().GetSecretAccessKeyOr("")
	creds := credentials.NewStaticCredentialsProvider(hmacAccessKey, hmacSecretKey, "")
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(config.DefaultObjectStorageRegion),
		awsconfig.WithCredentialsProvider(creds),
		awsconfig.WithBaseEndpoint(config.DefaultObjectStorageEndpoint),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("load AWS config: %w", err)
	}

	awsClient := s3.NewFromConfig(awsConfig)

	cleanup := func() {
		_ = hmacKeys.DeleteHmacKey(ctx, iamclient.DeleteHmacKeyRequest{
			ServiceAccount: serviceAccount,
			KeyName:        hmacKeyName,
		})
	}

	return awsClient, cleanup, nil
}
