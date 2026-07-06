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
	mock "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export/mock"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.uber.org/mock/gomock"
)

func TestStepDumpDiskImage_Run(t *testing.T) {
	t.Parallel()

	const checkAccessCommand = "qemu-img info /dev/disk/by-id/mws-disk-for-export"
	const dumpCommand = "qemu-img convert -f raw -O qcow2 -c /dev/disk/by-id/mws-disk-for-export image.qcow2"
	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tc := range []struct {
		name           string
		comm           packer.Communicator
		expectedAction multistep.StepAction
	}{
		{
			name: "ok",
			comm: func(t *testing.T) packer.Communicator {
				t.Helper()
				ctrl := gomock.NewController(t)
				comm := mock.NewMockCommunicator(ctrl)
				expectRemoteCmd(comm, checkAccessCommand, nil, 0)
				expectRemoteCmd(comm, dumpCommand, nil, 0)
				return comm
			}(t),
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "ok_with_sudo",
			comm: func(t *testing.T) packer.Communicator {
				t.Helper()
				ctrl := gomock.NewController(t)
				comm := mock.NewMockCommunicator(ctrl)
				expectRemoteCmd(comm, checkAccessCommand, nil, 1)
				expectRemoteCmd(comm, "sudo "+dumpCommand, nil, 0)
				return comm
			}(t),
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "check_access_error",
			comm: func(t *testing.T) packer.Communicator {
				t.Helper()
				ctrl := gomock.NewController(t)
				comm := mock.NewMockCommunicator(ctrl)
				expectRemoteCmd(comm, checkAccessCommand, errInternal, 0)
				return comm
			}(t),
			expectedAction: multistep.ActionHalt,
		},
		{
			name: "dump_error",
			comm: func(t *testing.T) packer.Communicator {
				t.Helper()
				ctrl := gomock.NewController(t)
				comm := mock.NewMockCommunicator(ctrl)
				expectRemoteCmd(comm, checkAccessCommand, nil, 0)
				expectRemoteCmd(comm, dumpCommand, errInternal, 0)
				return comm
			}(t),
			expectedAction: multistep.ActionHalt,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			state := new(multistep.BasicStateBag)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(mws.UIKey, ui)
			state.Put(mws.CommunicatorKey, tc.comm)

			step := &mwsexport.StepDumpDiskImage{}

			action := step.Run(t.Context(), state)
			require.Equal(t, tc.expectedAction, action)
			dir.String(t, tc.name+".out", writer.String())
		})
	}
}
