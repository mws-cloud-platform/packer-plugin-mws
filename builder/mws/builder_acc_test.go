// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	"github.com/hashicorp/packer-plugin-sdk/random"
	"github.com/mws-cloud-platform/packer-plugin-mws/example"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/go-sdk/mws"
	"go.mws.cloud/go-sdk/service/compute/client"
	computesdk "go.mws.cloud/go-sdk/service/compute/sdk"
)

func TestAccMWSBuilder(t *testing.T) {
	ctx := t.Context()
	sdk, err := mws.Load(ctx)
	require.NoError(t, err)
	imageClient, err := computesdk.NewImage(ctx, sdk)
	require.NoError(t, err)

	imageName := fmt.Sprintf("packer-acctest-%s-image", random.AlphaNumLower(6))

	testCase := &acctest.PluginTestCase{
		Name:           "builder_example",
		Template:       example.BuilderHCL,
		Type:           "mws",
		BuildExtraArgs: []string{"-var", "image_name=" + imageName},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil && buildCommand.ProcessState.ExitCode() != 0 {
				return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
			}
			if _, err := imageClient.GetImage(ctx, client.GetImageRequest{
				Image: imageName,
			}, client.WithWait()); err != nil {
				return fmt.Errorf("get image: %w", err)
			}
			return nil
		},
		Teardown: func() error {
			if err := imageClient.DeleteImage(ctx, client.DeleteImageRequest{
				Image: imageName,
			}, client.WithWait()); err != nil {
				return fmt.Errorf("delete image: %w", err)
			}
			return nil
		},
	}

	acctest.TestPlugin(t, testCase)
}
