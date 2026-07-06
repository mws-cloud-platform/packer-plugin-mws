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
	"github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export/mock"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.uber.org/mock/gomock"
)

const (
	hmacKeyName    = prefix + "hmac-key"
	serviceAccount = "sa"
)

func TestStepCreateHMACKey_Run(t *testing.T) {
	t.Parallel()

	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "ok"},
		{name: "error", err: errInternal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			state := new(multistep.BasicStateBag)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(mws.UIKey, ui)
			state.Put(mws.PrefixKey, prefix)
			driver := mock.NewMockDriver(ctrl)
			driver.EXPECT().
				CreateHMACKey(gomock.Any(), serviceAccount, hmacKeyName).
				Return("accessKey", "secretKey", tc.err)
			state.Put(mws.DriverKey, driver)

			step := &mwsexport.StepCreateHMACKey{
				ServiceAccount: serviceAccount,
				CleanupTimeout: cleanupTimeout,
			}

			action := step.Run(t.Context(), state)
			if tc.err == nil {
				require.Equal(t, multistep.ActionContinue, action)
			} else {
				require.Equal(t, multistep.ActionHalt, action)
				require.ErrorIs(t, state.Get(mws.ErrorKey).(error), tc.err)
			}
			dir.String(t, tc.name+".out", writer.String())
		})
	}
}

func TestStepCreateHMACKey_Cleanup(t *testing.T) {
	t.Parallel()

	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tc := range []struct {
		name string
		err  error
	}{
		{name: "ok"},
		{name: "error", err: errInternal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			state := new(multistep.BasicStateBag)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(mws.UIKey, ui)
			state.Put(mws.PrefixKey, prefix)
			driver := mock.NewMockDriver(ctrl)
			driver.EXPECT().
				DeleteHMACKey(gomock.Any(), serviceAccount, hmacKeyName).
				Return(tc.err)
			state.Put(mws.DriverKey, driver)

			step := &mwsexport.StepCreateHMACKey{
				ServiceAccount: serviceAccount,
				CleanupTimeout: cleanupTimeout,
			}

			step.Cleanup(state)
			dir.String(t, tc.name+".out", writer.String())
		})
	}
}
