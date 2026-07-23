// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"net"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/go-sdk/pkg/apimodels/ipaddress"
)

func TestConvertToIPv6(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ipv4          string
		prefix        string
		expectErr     bool
		expectedIPv6  string
		expectedError string
	}{
		{
			name:         "valid_ipv4_with_32_prefix",
			ipv4:         "192.0.2.33",
			prefix:       "64:ff9b::/32",
			expectErr:    false,
			expectedIPv6: "64:ff9b:c000:221::",
		},
		{
			name:         "valid_ipv4_with_40_prefix",
			ipv4:         "192.0.2.33",
			prefix:       "64:ff9b:1::/40",
			expectErr:    false,
			expectedIPv6: "64:ff9b:c0:2:2121::",
		},
		{
			name:         "valid_ipv4_with_48_prefix",
			ipv4:         "192.0.2.33",
			prefix:       "64:ff9b:1:2::/48",
			expectErr:    false,
			expectedIPv6: "64:ff9b:1:c000:202:2100::",
		},
		{
			name:         "valid_ipv4_with_56_prefix",
			ipv4:         "192.0.2.33",
			prefix:       "64:ff9b:1:2:3::/56",
			expectErr:    false,
			expectedIPv6: "64:ff9b:1:c0:0:221::",
		},
		{
			name:         "valid_ipv4_with_64_prefix",
			ipv4:         "192.0.2.33",
			prefix:       "64:ff9b:1:2:3:4::/64",
			expectErr:    false,
			expectedIPv6: "64:ff9b:1:2:c0c0:2:2100:0",
		},
		{
			name:         "valid_ipv4_with_96_prefix",
			ipv4:         "192.0.2.33",
			prefix:       "64:ff9b::/96",
			expectErr:    false,
			expectedIPv6: "64:ff9b::c000:221",
		},
		{
			name:          "nil_ipv4",
			ipv4:          "",
			prefix:        "64:ff9b::/32",
			expectErr:     true,
			expectedError: "ipv4 is nil",
		},
		{
			name:          "invalid_ipv4",
			ipv4:          "not.an.ip.address",
			prefix:        "64:ff9b::/32",
			expectErr:     true,
			expectedError: "ipv4 is nil",
		},
		{
			name:          "ipv6_instead_of_ipv4",
			ipv4:          "::1",
			prefix:        "64:ff9b::/32",
			expectErr:     true,
			expectedError: "invalid IPv4 address",
		},
		{
			name:          "invalid_prefix",
			ipv4:          "192.0.2.33",
			prefix:        "invalid-prefix",
			expectErr:     true,
			expectedError: "parse NAT64 prefix: invalid CIDR address: invalid-prefix",
		},
		{
			name:          "unsupported_prefix_length",
			ipv4:          "192.0.2.33",
			prefix:        "64:ff9b:1:2::/44",
			expectErr:     true,
			expectedError: "convert with RFC6052: unsupported RFC 6052 prefix length: 44",
		},
		{
			name:         "ipv4_loopback",
			ipv4:         "127.0.0.1",
			prefix:       "64:ff9b::/32",
			expectErr:    false,
			expectedIPv6: "64:ff9b:7f00:1::",
		},
		{
			name:         "ipv4_private_class_a",
			ipv4:         "10.0.0.1",
			prefix:       "64:ff9b::/32",
			expectErr:    false,
			expectedIPv6: "64:ff9b:a00:1::",
		},
		{
			name:         "ipv4_private_class_b",
			ipv4:         "172.16.0.1",
			prefix:       "64:ff9b::/32",
			expectErr:    false,
			expectedIPv6: "64:ff9b:ac10:1::",
		},
		{
			name:         "ipv4_private_class_c",
			ipv4:         "192.168.0.1",
			prefix:       "64:ff9b::/32",
			expectErr:    false,
			expectedIPv6: "64:ff9b:c0a8:1::",
		},
		{
			name:         "ipv4_broadcast",
			ipv4:         "255.255.255.255",
			prefix:       "64:ff9b::/32",
			expectErr:    false,
			expectedIPv6: "64:ff9b:ffff:ffff::",
		},
		{
			name:         "ipv4_zero_address",
			ipv4:         "0.0.0.0",
			prefix:       "64:ff9b::/32",
			expectErr:    false,
			expectedIPv6: "64:ff9b::",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ipv4 net.IP
			if tt.ipv4 != "" {
				ipv4 = net.ParseIP(tt.ipv4)
			}

			address, _ := ipaddress.NewIPAddress(ipv4)
			result, err := mws.ConvertToIPv6(new(address), tt.prefix)

			if tt.expectErr {
				require.Error(t, err)
				require.Equal(t, tt.expectedError, err.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)

				// Check that the result is a valid IPv6 address
				require.NotNil(t, result.ToNetIP())
				require.Equal(t, net.IPv6len, len(result.ToNetIP()))

				// Check that the result matches the expected IPv6 address
				require.Equal(t, tt.expectedIPv6, result.String())
			}
		})
	}
}
