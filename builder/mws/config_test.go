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
			name: "valid_basic_config_with_source_image",
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
			name: "valid_basic_config_with_source_snapshot",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_snapshot":      "test-snapshot",
					"use_external_address": true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full_config",
			raws: []any{
				map[string]any{
					"project":                             "test-project",
					"zone":                                "ru-central1-b",
					"base_endpoint":                       "https://custom.api.mwsapis.ru",
					"service_account_authorized_key_path": "/path/to/key",
					"virtual_machine_name":                "test-vm",
					"vm_type":                             "gen-2-16",
					"image_name":                          "test-image",
					"image_display_name":                  "Custom display name",
					"image_description":                   "Custom image description",
					"disk_name":                           "test-disk",
					"disk_type":                           "nbs-pl3",
					"disk_size":                           "50 GB",
					"disk_iops":                           int64(2000),
					"source_project":                      "source-project",
					"source_image":                        "source-image",
					"network_name":                        "test-network",
					"subnet_name":                         "test-subnet",
					"subnet_cidr":                         "10.0.0.0/8",
					"external_address_name":               "test-external-address",
					"nat64_enable":                        true,
					"cleanup_timeout":                     "2h",
					"use_external_address":                true,
				},
			},
			wantErr: false,
		},
		{
			name: "valid_config_with_use_external_address_false_and_defaults",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_image":         "test-image",
					"network_name":         "test-network",
					"subnet_name":          "test-subnet",
					"use_external_address": false,
				},
			},
			wantErr: false,
		},
		{
			name: "use_external_address_false_without_subnet_error",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_image":         "test-image",
					"use_external_address": false,
				},
			},
			wantErr: true,
		},
		{
			name: "use_external_address_false_with_external_address_error",
			raws: []any{
				map[string]any{
					"project":               "test-project",
					"source_image":          "test-image",
					"network_name":          "test-network",
					"subnet_name":           "test-subnet",
					"external_address_name": "test-external-address",
					"use_external_address":  false,
				},
			},
			wantErr: true,
		},
		{
			name: "missing_project_error",
			raws: []any{
				map[string]any{
					"source_image":         "test-image",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
		{
			name: "both_source_fields_error",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_image":         "test-image",
					"source_snapshot":      "test-snapshot",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
		{
			name: "no_source_fields_error",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_disk_size_error",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_image":         "test-image",
					"disk_size":            "invalid-size",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_subnet_CIDR_error",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_image":         "test-image",
					"subnet_cidr":          "invalid-cidr",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
		{
			name: "subnet_without_network_error",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_image":         "test-image",
					"subnet_name":          "test-subnet",
					"use_external_address": true,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_cleanup_timeout_error",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_snapshot":      "test-snapshot",
					"cleanup_timeout":      "invalid-duration",
					"use_external_address": true,
				},
			},
			wantErr: true,
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
