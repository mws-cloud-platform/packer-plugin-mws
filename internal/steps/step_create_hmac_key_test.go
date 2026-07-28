// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps_test

import (
	"bytes"
	"path"
	"testing"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/steps"
	mocksteps "github.com/mws-cloud-platform/packer-plugin-mws/internal/steps/mock"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.uber.org/mock/gomock"
)

const (
	hmacKeyName    = packerPrefix + "hmac-key"
	serviceAccount = "sa"
	cleanupTimeout = time.Hour
)

func TestStepCreateHMACKey_Run(t *testing.T) {
	t.Parallel()

	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tc := range []struct {
		name           string
		err            error
		serviceAccount string
		accessKey      string
		secretKey      string
	}{
		{name: "ok", serviceAccount: serviceAccount},
		{name: "error", err: errInternal, serviceAccount: serviceAccount},
		{name: "unset-service-account", accessKey: "test-access-key", secretKey: "test-secret-key"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			state := new(multistep.BasicStateBag)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(common.UIKey, ui)
			state.Put(common.PrefixKey, packerPrefix)
			driver := mocksteps.NewMockStepCreateHMACKeyDriver(ctrl)
			if tc.serviceAccount != "" {
				driver.EXPECT().
					CreateHMACKey(gomock.Any(), tc.serviceAccount, hmacKeyName).
					Return("accessKey", "secretKey", tc.err)
			}
			state.Put(common.DriverKey, driver)

			step := &steps.StepCreateHMACKey{
				ServiceAccount: tc.serviceAccount,
				AccessKey:      tc.accessKey,
				SecretKey:      tc.secretKey,
				CleanupTimeout: cleanupTimeout,
			}

			action := step.Run(t.Context(), state)
			if tc.err == nil {
				require.Equal(t, multistep.ActionContinue, action)
				if tc.serviceAccount != "" {
					require.Equal(t, "accessKey", state.Get(common.HMACAccessKeyStateKey))
					require.Equal(t, "secretKey", state.Get(common.HMACSecretKeyStateKey))
				} else {
					require.Equal(t, tc.accessKey, state.Get(common.HMACAccessKeyStateKey))
					require.Equal(t, tc.secretKey, state.Get(common.HMACSecretKeyStateKey))
				}
			} else {
				require.Equal(t, multistep.ActionHalt, action)
				require.ErrorIs(t, state.Get(common.ErrorKey).(error), tc.err)
			}
			dir.String(t, tc.name+".out", writer.String())
		})
	}
}

func TestStepCreateHMACKey_Cleanup(t *testing.T) {
	t.Parallel()

	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tc := range []struct {
		name           string
		err            error
		serviceAccount string
	}{
		{name: "ok", serviceAccount: serviceAccount},
		{name: "error", err: errInternal, serviceAccount: serviceAccount},
		{name: "unset-service-account", serviceAccount: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			state := new(multistep.BasicStateBag)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(common.UIKey, ui)
			state.Put(common.PrefixKey, packerPrefix)

			driver := mocksteps.NewMockStepCreateHMACKeyDriver(ctrl)
			if tc.serviceAccount != "" {
				driver.EXPECT().
					DeleteHMACKey(gomock.Any(), tc.serviceAccount, hmacKeyName).
					Return(tc.err)
				state.Put(common.HMACAccessKeyStateKey, "accessKey")
				state.Put(common.HMACSecretKeyStateKey, "secretKey")
			}
			state.Put(common.DriverKey, driver)

			step := &steps.StepCreateHMACKey{
				ServiceAccount: tc.serviceAccount,
				CleanupTimeout: cleanupTimeout,
			}

			step.Cleanup(state)
			dir.String(t, tc.name+".out", writer.String())
		})
	}
}
