// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config_test

import (
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"

	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestObjectStorageConfig(t *testing.T) {
	t.Parallel()
	tests := []ConfigTestCase{
		{
			name: "valid_basic_service_account_authentication",
			raws: []any{
				map[string]any{
					"service_account": "test-service-account",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_basic_access_key_secret_key_authentication",
			raws: []any{
				map[string]any{
					"access_key": "test-access-key",
					"secret_key": "test-secret-key",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full",
			raws: []any{
				map[string]any{
					"service_account":         "test-service-account",
					"object_storage_endpoint": "https://custom.api.mwsapis.ru",
					"object_storage_region":   "ru-central2",
				},
			},
			wantErr: false,
		},
		{
			name: "error_access_key_without_secret_key",
			raws: []any{
				map[string]any{
					"access_key": "test-access-key",
				},
			},
			wantErr: true,
		},
		{
			name: "error_secret_key_without_access_key",
			raws: []any{
				map[string]any{
					"secret_key": "test-secret-key",
				},
			},
			wantErr: true,
		},
		{
			name: "error_missing_authentication",
			raws: []any{
				map[string]any{},
			},
			wantErr: true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, tt.ConfigTest(&config.ObjectStorageConfig{}, expectedDir))
	}
}
