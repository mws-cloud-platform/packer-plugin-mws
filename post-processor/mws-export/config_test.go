// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport_test

import (
	"path"
	"testing"

	mwsexport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-export"
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
					"service_account":      "test-service-account",
					"object_storage_path":  "test-bucket/path/to/image.qcow2",
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
					// VirtualMachineConfig
					"virtual_machine_name": "test-vm",
					"vm_type":              "gen-2-16",
					"cloud_config":         "#cloud-config\npackages:\n  - nginx",
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
					// DiskForExportConfig
					"disk_for_export_type":     "nbs-pl3",
					"disk_for_export_iops":     int64(2000),
					"image_for_export_project": "export-image-project",
					"image_for_export":         "export-image",
					// ObjectStorageConfig
					"service_account":         "test-service-account",
					"object_storage_endpoint": "https://custom.api.mwsapis.ru",
					"object_storage_region":   "ru-central2",

					"object_storage_path": "test-bucket/path/to/image.qcow2",
				},
			},
			wantErr: false,
		},
		{
			name: "error_missing_object_storage_path",
			raws: []any{
				map[string]any{
					"project":              "test-project",
					"source_image":         "test-image",
					"use_external_address": true,
					"service_account":      "test-service-account",
				},
			},
			wantErr: true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &mwsexport.Config{}
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
