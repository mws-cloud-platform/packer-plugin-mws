// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config_test

import (
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"

	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestImageConfig(t *testing.T) {
	t.Parallel()
	tests := []ConfigTestCase{
		{
			name: "valid_basic",
			raws: []any{
				map[string]any{},
			},
			wantErr: false,
		},
		{
			name: "valid_full",
			raws: []any{
				map[string]any{
					"image_name":         "test-image-name",
					"image_display_name": "Test image display name",
					"image_description":  "Test image description.",
				},
			},
			wantErr: false,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, tt.ConfigTest(&config.ImageConfig{}, expectedDir))
	}
}
