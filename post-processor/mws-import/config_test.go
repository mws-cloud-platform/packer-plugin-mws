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
			name: "valid_basic_config_with_service_account",
			raws: []any{
				map[string]any{
					"project":             "test-project",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"service_account":     "test-service-account",
					"image_name":          "test-image",
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
					"image_name":                          "test-image",
					"image_display_name":                  "Custom display name",
					"image_description":                   "Custom image description",
					"cleanup_timeout":                     "2h",
					"object_storage_path":                 "test-bucket/path/to/image.qcow2",
					"object_storage_endpoint":             "https://custom.storage.mwsapis.ru",
					"object_storage_region":               "ru-central1-a",
					"service_account":                     "test-service-account",
				},
			},
			wantErr: false,
		},
		{
			name: "missing_project_error",
			raws: []any{
				map[string]any{
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"service_account":     "test-service-account",
					"image_name":          "test-image",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_cleanup_timeout_error",
			raws: []any{
				map[string]any{
					"project":             "test-project",
					"cleanup_timeout":     "invalid-duration",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"service_account":     "test-service-account",
					"image_name":          "test-image",
				},
			},
			wantErr: true,
		},
		{
			name: "missing_object_storage_authentication_error",
			raws: []any{
				map[string]any{
					"project":             "test-project",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"image_name":          "test-image",
				},
			},
			wantErr: true,
		},
		{
			name: "valid_service_account_authentication",
			raws: []any{
				map[string]any{
					"project":             "test-project",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"service_account":     "test-service-account",
					"image_name":          "test-image",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_access_key_secret_key_authentication",
			raws: []any{
				map[string]any{
					"project":             "test-project",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"access_key":          "test-access-key",
					"secret_key":          "test-secret-key",
					"image_name":          "test-image",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid_access_key_without_secret_key_error",
			raws: []any{
				map[string]any{
					"project":             "test-project",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"access_key":          "test-access-key",
					"image_name":          "test-image",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid_secret_key_without_access_key_error",
			raws: []any{
				map[string]any{
					"project":             "test-project",
					"object_storage_path": "test-bucket/path/to/image.qcow2",
					"secret_key":          "test-secret-key",
					"image_name":          "test-image",
				},
			},
			wantErr: true,
		},
		{
			name: "missing_object_storage_path_error",
			raws: []any{
				map[string]any{
					"project":         "test-project",
					"service_account": "test-service-account",
					"image_name":      "test-image",
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
