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
	mwsexport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestStepUploadImage_Run(t *testing.T) {
	t.Parallel()
	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	state := new(multistep.BasicStateBag)
	writer := new(bytes.Buffer)
	ui := &packer.BasicUi{Writer: writer}
	state.Put(mws.UIKey, ui)
	comm := &packer.MockCommunicator{}
	state.Put(mws.CommunicatorKey, comm)

	step := &mwsexport.StepUploadImage{
		Region:   "ru-central1",
		Endpoint: "storage.mwsapis.ru",
		Path:     "bucket/path/to/image",
	}

	action := step.Run(t.Context(), state)
	require.Equal(t, multistep.ActionContinue, action)
	require.True(t, comm.StartCalled)
	dir.String(t, t.Name()+"_command.txt", comm.StartCmd.Command)
	dir.String(t, t.Name()+".out", writer.String())
}
