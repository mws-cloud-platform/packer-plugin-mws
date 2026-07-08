// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package cloudinit_test

import (
	"path"
	"testing"

	"github.com/mws-cloud-platform/packer-plugin-mws/internal/cloudinit"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestPrepareCloudConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		sshUsername       string
		sshPublicKey      string
		customCloudConfig string
		expectError       bool
	}{
		{
			name:              "NoCustomConfig",
			sshUsername:       "packer",
			sshPublicKey:      "key",
			customCloudConfig: "",
		},
		{
			name:         "CustomConfigWithoutUsers",
			sshUsername:  "packer",
			sshPublicKey: "key",
			customCloudConfig: `#cloud-config
packages:
  - nginx
runcmd:
  - echo "Hello World"`,
		},
		{
			name:         "CustomConfigWithUsers",
			sshUsername:  "packer",
			sshPublicKey: "key",
			customCloudConfig: `#cloud-config
users:
  - name: custom
    groups: docker
    shell: /bin/zsh
packages:
  - nginx`,
		},
		{
			name:         "CustomConfigWithUsersObject",
			sshUsername:  "packer",
			sshPublicKey: "key",
			customCloudConfig: `#cloud-config
users:
  name: custom
  groups: docker
  shell: /bin/zsh
packages:
  - nginx`,
		},
		{
			name:         "CustomConfigWithUsersString",
			sshUsername:  "packer",
			sshPublicKey: "key",
			customCloudConfig: `#cloud-config
users: custom
packages:
  - nginx`,
		},
		{
			name:              "InvalidCustomCloudConfig",
			sshUsername:       "packer",
			sshPublicKey:      "key",
			customCloudConfig: `invalid: yaml: content:`,
			expectError:       true,
		},
	}

	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual, err := cloudinit.PrepareCloudConfig(tt.sshUsername, tt.sshPublicKey, tt.customCloudConfig)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				dir.String(t, tt.name+".yaml", actual)
			}
		})
	}
}
