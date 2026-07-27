// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config_test

import (
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"

	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestAccessConfig(t *testing.T) {
	t.Parallel()
	tests := []ConfigTestCase{
		{
			name: "valid_basic",
			raws: []any{
				map[string]any{
					"project": "test-project",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full_service_account_auth",
			raws: []any{
				map[string]any{
					"project":                             "test-project",
					"zone":                                "ru-central1-b",
					"base_endpoint":                       "https://custom.api.mwsapis.ru",
					"service_account_authorized_key_path": "/path/to/key",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full_token_auth",
			raws: []any{
				map[string]any{
					"project":       "test-project",
					"zone":          "ru-central1-b",
					"base_endpoint": "https://custom.api.mwsapis.ru",
					"token":         "secret-token",
				},
			},
			wantErr: false,
		},
		{
			name: "error_missing_project",
			raws: []any{
				map[string]any{},
			},
			wantErr: true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, ConfigTest(&config.AccessConfig{}, tt, expectedDir))
	}
}
