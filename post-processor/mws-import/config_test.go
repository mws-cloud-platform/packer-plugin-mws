// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport_test

import (
	"path"
	"testing"

	mwsimport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-import"
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
					"project":             "test-project",
					"service_account":     "test-service-account",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
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
					// ObjectStorageConfig
					"service_account":         "test-service-account",
					"object_storage_endpoint": "https://custom.api.mwsapis.ru",
					"object_storage_region":   "ru-central2",

					"object_storage_path": "test-bucket/path/to/image.qcow2",

					"cleanup_timeout": "2h",
				},
			},
			wantErr: false,
		},
		{
			name: "error_missing_object_storage_path",
			raws: []any{
				map[string]any{
					"project":         "test-project",
					"service_account": "test-service-account",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_cleanup_timeout_error",
			raws: []any{
				map[string]any{
					"project":             "test-project",
					"service_account":     "test-service-account",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"cleanup_timeout":     "invalid-duration",
				},
			},
			wantErr: true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &mwsimport.Config{}
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
