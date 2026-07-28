// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config_test

import (
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"

	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestNetworkConfig(t *testing.T) {
	t.Parallel()
	tests := []ConfigTestCase{
		{
			name: "valid_basic_use_external_address_true",
			raws: []any{
				map[string]any{
					"use_external_address": true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid_basic_use_external_address_false",
			raws: []any{
				map[string]any{
					"network_name":         "test-network",
					"subnet_name":          "test-subnet",
					"use_external_address": false,
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full",
			raws: []any{
				map[string]any{
					"network_name":          "test-network",
					"subnet_name":           "test-subnet",
					"subnet_cidr":           "10.0.0.0/8",
					"use_external_address":  true,
					"external_address_name": "test-external-address",
					"nat64_enable":          true,
					"nat64_ipv6_prefix":     "2a02:5501:0:6000::/64",
				},
			},
			wantErr: false,
		},

		{
			name: "error_use_external_address_false_without_subnet",
			raws: []any{
				map[string]any{
					"use_external_address": false,
				},
			},
			wantErr: true,
		},
		{
			name: "error_subnet_without_network",
			raws: []any{
				map[string]any{
					"subnet_name":          "test-subnet",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
		{
			name: "error_use_external_address_false_with_external_address",
			raws: []any{
				map[string]any{
					"network_name":          "test-network",
					"subnet_name":           "test-subnet",
					"use_external_address":  false,
					"external_address_name": "test-external-address",
				},
			},
			wantErr: true,
		},
		{
			name: "error_invalid_subnet_cidr",
			raws: []any{
				map[string]any{
					"subnet_cidr":          "invalid-cidr",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
		{
			name: "error_invalid_nat64_ipv6_prefix",
			raws: []any{
				map[string]any{
					"nat64_ipv6_prefix":    "invalid-nat64-ipv6-prefix",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, tt.ConfigTest(&config.NetworkConfig{}, expectedDir))
	}
}
