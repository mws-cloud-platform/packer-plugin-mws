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

func TestMergeCloudInit(t *testing.T) {
	tests := []struct {
		name         string
		baseUser     string
		baseSSHKey   string
		customConfig string
		expectError  bool
	}{
		{
			name:         "NoCustomConfig",
			baseUser:     "testuser",
			baseSSHKey:   "testkey",
			customConfig: "",
		},
		{
			name:       "CustomConfigWithoutUsers",
			baseUser:   "testuser",
			baseSSHKey: "testkey",
			customConfig: `packages:
  - nginx
runcmd:
  - echo "Hello World"`,
		},
		{
			name:       "CustomConfigWithUsers",
			baseUser:   "testuser",
			baseSSHKey: "testkey",
			customConfig: `users:
  - name: customuser
    groups: docker
    shell: /bin/zsh
packages:
  - nginx`,
		},
		{
			name:         "InvalidYAML",
			baseUser:     "testuser",
			baseSSHKey:   "testkey",
			customConfig: `invalid: yaml: content:`,
			expectError:  true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mws.MergeCloudInit(tt.baseUser, tt.baseSSHKey, tt.customConfig)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				expectedDir.String(t, tt.name+".yml", result)
			}
		})
	}
}
