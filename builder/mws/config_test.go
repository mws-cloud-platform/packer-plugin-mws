// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestConfig_Prepare(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raws    []any
		wantErr bool
	}{
		{
			name: "valid_basic",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_image":         "test-image",
					"use_external_address": true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full",
			raws: []any{
				map[string]any{
					// AccessConfig
					"project":                             "test-project",
					"zone":                                "ru-central1-b",
					"base_endpoint":                       "https://custom.api.mwsapis.ru",
					"service_account_authorized_key_path": "/path/to/key",
					// ImageConfig
					"image_name":         "test-image-name",
					"image_display_name": "Test image display name",
					"image_description":  "Test image description.",
					// VirtualMachineConfig
					"virtual_machine_name":    "test-vm",
					"vm_type":                 "gen-2-16",
					"cloud_config":            "#cloud-config\npackages:\n  - nginx",
					"vm_service_account":      "test-admin",
					"serial_console_log_file": "serial_console.log",
					"cleanup_timeout":         "2h",
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
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &mws.Config{}
			err := c.Prepare(tt.raws...)

			if tt.wantErr {
				require.Error(t, err)
				expectedDir.String(t, tt.name+".txt", err.Error())
			} else {
				require.NoError(t, err)
				expectedDir.JSON(t, tt.name+".json", c)
			}
		})
	}
}
