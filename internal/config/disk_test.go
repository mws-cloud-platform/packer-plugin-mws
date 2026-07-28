// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config_test

import (
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"

	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestDiskConfig(t *testing.T) {
	t.Parallel()
	tests := []ConfigTestCase{
		{
			name: "valid_basic_with_source_image",
			raws: []any{
				map[string]any{
					"source_image": "source-image",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_basic_with_source_snapshot",
			raws: []any{
				map[string]any{
					"source_snapshot": "source-snapshot",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full_with_source_image",
			raws: []any{
				map[string]any{
					"disk_name":      "test-disk",
					"disk_type":      "nbs-pl3",
					"disk_size":      "50 GB",
					"disk_iops":      int64(2000),
					"source_project": "source-project",
					"source_image":   "source-image",
				},
			},
			wantErr: false,
		},
		{
			name: "valid_full_with_source_snapshot",
			raws: []any{
				map[string]any{
					"disk_name":       "test-disk",
					"disk_type":       "nbs-pl3",
					"disk_size":       "50 GB",
					"disk_iops":       int64(2000),
					"source_project":  "source-project",
					"source_snapshot": "source-snapshot",
				},
			},
			wantErr: false,
		},
		{
			name: "error_missing_source",
			raws: []any{
				map[string]any{},
			},
			wantErr: true,
		},
		{
			name: "error_both_source",
			raws: []any{
				map[string]any{
					"source_image":    "test-image",
					"source_snapshot": "test-snapshot",
				},
			},
			wantErr: true,
		},
		{
			name: "error_invalid_disk_size",
			raws: []any{
				map[string]any{
					"disk_size":    "invalid-size",
					"source_image": "test-image",
				},
			},
			wantErr: true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, tt.ConfigTest(&config.DiskConfig{}, expectedDir))
	}
}
