// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

func CommHost(host string) func(state multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		ui := state.Get(UIKey).(packer.Ui)
		if host != "" {
			ui.Sayf("Using provided ssh host %q", host)
			return host, nil
		}

		ipAddress, ok := state.Get(InstanceIPKey).(string)
		if !ok {
			return "", consterr.Error("instance IP address not found")
		}
		return ipAddress, nil
	}
}
