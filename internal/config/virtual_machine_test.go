// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config_test

import (
	_ "embed"
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"

	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

//go:embed testdata/cloud-config.yaml
var testCloudConfig string

func TestVirtualMachineConfig(t *testing.T) {
	t.Parallel()
	tests := []ConfigTestCase{
		{
			name: "valid_basic",
			raws: []any{
				map[string]any{
					"source_image":         "source-image",
					"use_external_address": true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full",
			raws: []any{
				map[string]any{
					"virtual_machine_name": "test-vm",
					"vm_type":              "gen-2-16",
					"cloud_config":         testCloudConfig,
					"vm_service_account":   "test-admin",
					"cleanup_timeout":      "2h",
					// DiskConfig
					"disk_name":      "test-disk",
					"disk_type":      "nbs-pl3",
					"disk_size":      "50 GB",
					"disk_iops":      int64(2000),
					"source_project": "source-project",
					"source_image":   "source-image",
					// NetworkConfig
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
			name: "error_invalid_cleanup_timeout",
			raws: []any{
				map[string]any{
					"cleanup_timeout":      "invalid-duration",
					"source_image":         "source-image",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, tt.ConfigTest(&config.VirtualMachineConfig{}, expectedDir))
	}
}
