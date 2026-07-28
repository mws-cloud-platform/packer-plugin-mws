// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"context"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/common"
	drivermws "github.com/mws-cloud-platform/packer-plugin-mws/internal/driver"
)

type StepCreateHMACKey struct {
	ServiceAccount string
	AccessKey      string
	SecretKey      string
	CleanupTimeout time.Duration
}

func (s *StepCreateHMACKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(common.DriverKey).(StepCreateHMACKeyDriver)
	ui := state.Get(common.UIKey).(packer.Ui)
	name := s.hmacKeyName(state)

	accessKey := s.AccessKey
	secretKey := s.SecretKey

	if s.ServiceAccount == "" {
		ui.Say("Using provided HMAC key...")
	} else {
		ui.Say("Creating temporary HMAC key...")

		var err error
		accessKey, secretKey, err = driver.CreateHMACKey(ctx, s.ServiceAccount, name)
		if err != nil {
			return common.ActionHaltWithErrorf(state, "create hmac key: %w", err)
		}

		ui.Say("HMAC key created")
	}

	state.Put(common.HMACAccessKeyStateKey, accessKey)
	state.Put(common.HMACSecretKeyStateKey, secretKey)

	return multistep.ActionContinue
}

func (s *StepCreateHMACKey) Cleanup(state multistep.StateBag) {
	driver := state.Get(common.DriverKey).(StepCreateHMACKeyDriver)
	ui := state.Get(common.UIKey).(packer.Ui)

	if s.ServiceAccount == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.CleanupTimeout)
	defer cancel()

	drivermws.DeleteSubWithUI(ctx, ui, "HMAC key", s.hmacKeyName(state), s.ServiceAccount, driver.DeleteHMACKey)
}

func (s *StepCreateHMACKey) hmacKeyName(state multistep.StateBag) string {
	prefix := state.Get(common.PrefixKey).(string)
	return prefix + "hmac-key"
}
