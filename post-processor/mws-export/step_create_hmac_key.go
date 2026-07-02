// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"
	"time"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
)

type StepCreateHMACKey struct {
	ServiceAccount string
	CleanupTimeout time.Duration
}

func (s *StepCreateHMACKey) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(mws.DriverKey).(Driver)
	ui := state.Get(mws.UIKey).(packer.Ui)
	name := s.hmacKeyName(state)

	ui.Say("Creating temporary HMAC key...")

	accessKey, secretKey, err := driver.CreateHMACKey(ctx, s.ServiceAccount, name)
	if err != nil {
		return mws.ActionHaltWithErrorf(state, "create hmac key: %w", err)
	}

	ui.Say("HMAC key created")
	state.Put(HMACAccessKeyStateKey, accessKey)
	state.Put(HMACSecretKeyStateKey, secretKey)

	return multistep.ActionContinue
}

func (s *StepCreateHMACKey) Cleanup(state multistep.StateBag) {
	driver := state.Get(mws.DriverKey).(Driver)
	ui := state.Get(mws.UIKey).(packer.Ui)

	if _, ok := state.GetOk(HMACAccessKeyStateKey); !ok {
		ctx, cancel := context.WithTimeout(context.Background(), s.CleanupTimeout)
		defer cancel()

		name := s.hmacKeyName(state)

		ui.Say("Deleting HMAC key...")
		if err := driver.DeleteHMACKey(ctx, s.ServiceAccount, name); err != nil {
			ui.Errorf("Error deleting HMAC key %q. Please delete it manually.\n"+
				"Error: %v.", name, err)
		} else {
			ui.Sayf("HMAC key %q deleted", name)
		}
	}
}

func (s *StepCreateHMACKey) hmacKeyName(state multistep.StateBag) string {
	prefix := state.Get(mws.PrefixKey).(string)
	return prefix + "hmac-key"
}
