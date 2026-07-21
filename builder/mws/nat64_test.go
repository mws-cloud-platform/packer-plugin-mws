// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"net"
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestConvertToIPv6(t *testing.T) {
	t.Parallel()
	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	tests := []struct {
		name      string
		ipv4      string
		prefix    string
		expectErr bool
	}{
		{
			name:      "valid_ipv4_with_32_prefix",
			ipv4:      "192.0.2.33",
			prefix:    "64:ff9b::/32",
			expectErr: false,
		},
		{
			name:      "valid_ipv4_with_40_prefix",
			ipv4:      "192.0.2.33",
			prefix:    "64:ff9b:1::/40",
			expectErr: false,
		},
		{
			name:      "valid_ipv4_with_48_prefix",
			ipv4:      "192.0.2.33",
			prefix:    "64:ff9b:1:2::/48",
			expectErr: false,
		},
		{
			name:      "valid_ipv4_with_56_prefix",
			ipv4:      "192.0.2.33",
			prefix:    "64:ff9b:1:2:3::/56",
			expectErr: false,
		},
		{
			name:      "valid_ipv4_with_64_prefix",
			ipv4:      "192.0.2.33",
			prefix:    "64:ff9b:1:2:3:4::/64",
			expectErr: false,
		},
		{
			name:      "valid_ipv4_with_96_prefix",
			ipv4:      "192.0.2.33",
			prefix:    "64:ff9b::/96",
			expectErr: false,
		},
		{
			name:      "nil_ipv4",
			ipv4:      "",
			prefix:    "64:ff9b::/32",
			expectErr: true,
		},
		{
			name:      "invalid_ipv4",
			ipv4:      "not.an.ip.address",
			prefix:    "64:ff9b::/32",
			expectErr: true,
		},
		{
			name:      "ipv6_instead_of_ipv4",
			ipv4:      "::1",
			prefix:    "64:ff9b::/32",
			expectErr: true,
		},
		{
			name:      "invalid_prefix",
			ipv4:      "192.0.2.33",
			prefix:    "invalid-prefix",
			expectErr: true,
		},
		{
			name:      "unsupported_prefix_length",
			ipv4:      "192.0.2.33",
			prefix:    "64:ff9b:1:2::/44",
			expectErr: true,
		},
		{
			name:      "ipv4_loopback",
			ipv4:      "127.0.0.1",
			prefix:    "64:ff9b::/32",
			expectErr: false,
		},
		{
			name:      "ipv4_private_class_a",
			ipv4:      "10.0.0.1",
			prefix:    "64:ff9b::/32",
			expectErr: false,
		},
		{
			name:      "ipv4_private_class_b",
			ipv4:      "172.16.0.1",
			prefix:    "64:ff9b::/32",
			expectErr: false,
		},
		{
			name:      "ipv4_private_class_c",
			ipv4:      "192.168.0.1",
			prefix:    "64:ff9b::/32",
			expectErr: false,
		},
		{
			name:      "ipv4_broadcast",
			ipv4:      "255.255.255.255",
			prefix:    "64:ff9b::/32",
			expectErr: false,
		},
		{
			name:      "ipv4_zero_address",
			ipv4:      "0.0.0.0",
			prefix:    "64:ff9b::/32",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ipv4 net.IP
			if tt.ipv4 != "" {
				ipv4 = net.ParseIP(tt.ipv4)
			}

			result, err := mws.ConvertToIPv6(ipv4, tt.prefix)

			if tt.expectErr {
				require.Error(t, err)
				expectedDir.String(t, tt.name+".error", err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				// Check that the result is a valid IPv6 address
				require.NotNil(t, result.ToNetIP())
				require.Equal(t, net.IPv6len, len(result.ToNetIP()))

				// Save the string representation of the IPv6 address
				expectedDir.String(t, tt.name+".ipv6", result.String())
			}
		})
	}
}
