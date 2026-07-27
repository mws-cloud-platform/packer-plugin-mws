// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport_test

import (
	"bytes"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	mwsexport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestStepUploadAWSSharedCredsFile_Run(t *testing.T) {
	t.Parallel()
	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	state := new(multistep.BasicStateBag)
	writer := new(bytes.Buffer)
	ui := &packer.BasicUi{Writer: writer}
	state.Put(common.UIKey, ui)
	comm := &packer.MockCommunicator{}
	state.Put(common.CommunicatorKey, comm)
	state.Put(common.HMACAccessKeyStateKey, "access_key")
	state.Put(common.HMACSecretKeyStateKey, "secret_key")

	step := &mwsexport.StepUploadAWSSharedCredsFile{}

	action := step.Run(t.Context(), state)
	require.Equal(t, multistep.ActionContinue, action)
	require.True(t, comm.UploadCalled)
	require.Equal(t, "/tmp/aws_credentials", comm.UploadPath)
	dir.String(t, t.Name()+"_upload_data.txt", comm.UploadData)
	dir.String(t, t.Name()+".out", writer.String())
}
