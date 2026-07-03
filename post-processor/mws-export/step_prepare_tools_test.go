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

func TestStepPrepareTools_Run(t *testing.T) {
	t.Parallel()

	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tc := range []struct {
		name              string
		configureCommMock func(*testing.T, *mock.MockCommunicator) *mock.MockCommunicator
		expectedAction    multistep.StepAction
	}{
		{
			name: "already_installed",
			configureCommMock: func(t *testing.T, comm *mock.MockCommunicator) *mock.MockCommunicator {
				expectRemoteCmd(comm, "which qemu-img", nil, 0)
				expectRemoteCmd(comm, "which aws", nil, 0)
				return comm
			},
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "all_installed",
			configureCommMock: func(t *testing.T, comm *mock.MockCommunicator) *mock.MockCommunicator {
				expectRemoteCmd(comm, "which qemu-img", nil, 1)
				expectRemoteCmd(comm, "which aws", nil, 1)
				expectRemoteCmd(comm, "which apt", nil, 0)
				expectRemoteCmd(comm, "apt update", nil, 0)
				expectRemoteCmd(comm, "apt install -y qemu-utils", nil, 0)
				expectRemoteCmd(comm, "apt install -y awscli", nil, 0)
				return comm
			},
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "all_installed_with_sudo",
			configureCommMock: func(t *testing.T, comm *mock.MockCommunicator) *mock.MockCommunicator {
				expectRemoteCmd(comm, "which qemu-img", nil, 1)
				expectRemoteCmd(comm, "which aws", nil, 1)
				expectRemoteCmd(comm, "which apt", nil, 0)
				expectRemoteCmd(comm, "apt update", nil, 1)
				expectRemoteCmd(comm, "sudo apt update", nil, 0)
				expectRemoteCmd(comm, "apt install -y qemu-utils", nil, 1)
				expectRemoteCmd(comm, "sudo apt install -y qemu-utils", nil, 0)
				expectRemoteCmd(comm, "apt install -y awscli", nil, 1)
				expectRemoteCmd(comm, "sudo apt install -y awscli", nil, 0)
				return comm
			},
			expectedAction: multistep.ActionContinue,
		},
		{
			name: "error",
			configureCommMock: func(t *testing.T, comm *mock.MockCommunicator) *mock.MockCommunicator {
				expectRemoteCmd(comm, "which qemu-img", errInternal, 0)
				return comm
			},
			expectedAction: multistep.ActionHalt,
		},
		{
			name: "apt_not_found",
			configureCommMock: func(t *testing.T, comm *mock.MockCommunicator) *mock.MockCommunicator {
				expectRemoteCmd(comm, "which qemu-img", nil, 1)
				expectRemoteCmd(comm, "which aws", nil, 1)
				expectRemoteCmd(comm, "which apt", nil, 1)
				return comm
			},
			expectedAction: multistep.ActionHalt,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			state := new(multistep.BasicStateBag)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(mws.UIKey, ui)
			comm := mock.NewMockCommunicator(ctrl)
			state.Put(mws.CommunicatorKey, tc.configureCommMock(t, comm))

			step := &mwsexport.StepPrepareTools{}

			action := step.Run(t.Context(), state)
			require.Equal(t, tc.expectedAction, action)
			dir.String(t, tc.name+".out", writer.String())
		})
	}
}
