// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"bytes"
	"path"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	mockmws "github.com/mws-cloud-platform/packer-plugin-mws/builder/mws/mock"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/testutil"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.uber.org/mock/gomock"
)

func TestStepShutdownVirtualMachine(t *testing.T) {
	t.Parallel()
	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name               string
		virtualMachineName string
		prepare            func(*mockmws.MockDriver)
		expectedError      bool
	}{
		{
			name:               "success_with_default_name",
			virtualMachineName: "",
			prepare: func(driver *mockmws.MockDriver) {
				driver.EXPECT().
					ShutdownVirtualMachine(gomock.Any(), defaultVirtualMachineName).
					Return(nil).
					Times(1)
			},
		},
		{
			name:               "success_with_custom_name",
			virtualMachineName: "custom-vm",
			prepare: func(driver *mockmws.MockDriver) {
				driver.EXPECT().
					ShutdownVirtualMachine(gomock.Any(), "custom-vm").
					Return(nil).
					Times(1)
			},
		},
		{
			name:               "shutdown_error",
			virtualMachineName: "",
			prepare: func(driver *mockmws.MockDriver) {
				driver.EXPECT().
					ShutdownVirtualMachine(gomock.Any(), defaultVirtualMachineName).
					Return(errInternal).
					Times(1)
			},
			expectedError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			driver := mockmws.NewMockDriver(ctrl)

			state := new(multistep.BasicStateBag)
			state.Put(common.DriverKey, driver)
			state.Put(common.PrefixKey, packerPrefix)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(common.UIKey, ui)

			if tt.prepare != nil {
				tt.prepare(driver)
			}

			step := &mws.StepShutdownVirtualMachine{
				VirtualMachineName: tt.virtualMachineName,
			}

			action := step.Run(t.Context(), state)
			if tt.expectedError {
				testutil.RequireActionHalt(t, state, action)
			} else {
				testutil.RequireActionContinue(t, state, action)
			}
			dir.String(t, tt.name+".out", writer.String())
		})
	}
}
