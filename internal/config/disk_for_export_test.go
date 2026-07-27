// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config_test

import (
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/config"

	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestDiskForExportConfig(t *testing.T) {
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
					"disk_for_export_type":     "nbs-pl3",
					"disk_for_export_iops":     int64(2000),
					"image_for_export_project": "export-image-project",
					"image_for_export":         "export-image",
				},
			},
			wantErr: false,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, ConfigTest(&config.DiskForExportConfig{}, tt, expectedDir))
	}
}
