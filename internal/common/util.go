// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

const ErrUnexpected = consterr.Error("plugin unexpected error")

func ActionHaltWithErrorf(state multistep.StateBag, format string, args ...any) multistep.StepAction {
	return ActionHaltWithError(state, fmt.Errorf(format, args...))
}

func ActionHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get(UIKey).(packer.Ui)
	state.Put(ErrorKey, err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

func StateGetOkString(state multistep.StateBag, key string) string {
	if val, ok := state.GetOk(key); ok {
		return val.(string)
	}
	return ""
}
