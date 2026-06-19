// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildSSHScript(t *testing.T) {
	t.Parallel()

	userData := buildSSHScript("packer", "ssh-ed25519 test-key")

	require.Contains(t, userData, "  - name: packer\n")
	require.Contains(t, userData, "      - ssh-ed25519 test-key\n")
	require.Contains(t, userData, "s/[[:space:]]*(#.*)?$/ packer \\1/' /etc/ssh/sshd_config")
	require.NotContains(t, userData, "PermitRootLogin")
	require.Contains(t, userData, "reload-or-restart sshd.service || systemctl reload-or-restart ssh.service")
	require.Contains(t, userData, "if mountpoint -q /tmp; then mount -o remount,exec,nodev,nosuid /tmp; fi")
}
