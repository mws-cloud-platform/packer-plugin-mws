// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport_test

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	mock "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export/mock"
	"go.uber.org/mock/gomock"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -destination=mock/communicator_mock.go github.com/hashicorp/packer-plugin-sdk/packer Communicator

func expectRemoteCmd(comm *mock.MockCommunicator, command string, err error, exitCode int) {
	comm.EXPECT().
		Start(gomock.Any(), matchRemoteCmd(command)).
		DoAndReturn(func(_ context.Context, rc *packer.RemoteCmd) error {
			if err != nil {
				return err
			}
			rc.SetExited(exitCode)
			return nil
		})
}

func matchRemoteCmd(command string) *remoteCmdMatcher {
	return &remoteCmdMatcher{command: command}
}

type remoteCmdMatcher struct {
	command string
}

func (c *remoteCmdMatcher) Matches(x any) bool {
	actual, ok := x.(*packer.RemoteCmd)
	if !ok {
		return false
	}
	return actual.Command == c.command
}

func (c *remoteCmdMatcher) String() string {
	return fmt.Sprintf("matches remote command %q", c.command)
}
