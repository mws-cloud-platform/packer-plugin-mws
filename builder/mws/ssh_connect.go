// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

func CommHost(state multistep.StateBag) (string, error) {
	ipAddress, ok := state.Get(instanceIpKey).(string)
	if !ok {
		return "", consterr.Error("instance IP address not found")
	}
	return ipAddress, nil
}
