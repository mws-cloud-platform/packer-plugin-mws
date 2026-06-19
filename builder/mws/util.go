// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
)

func actionHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get(UIKey).(packer.Ui)
	state.Put(ErrorKey, err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func stateGetOkString(state multistep.StateBag, key string) string {
	if val, ok := state.GetOk(key); ok {
		return val.(string)
	}
	return ""
}
