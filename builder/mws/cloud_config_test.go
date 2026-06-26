// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

func TestCloudConfig_NewCloudConfig(t *testing.T) {
	tests := []struct {
		name              string
		customCloudConfig string
		expectError       bool
	}{
		{
			name:              "NoCustomConfig",
			customCloudConfig: "",
		},
		{
			name: "CustomConfigWithoutUsers",
			customCloudConfig: `#cloud-config
packages:
  - nginx
runcmd:
  - echo "Hello World"`,
		},
		{
			name: "CustomConfigWithUsers",
			customCloudConfig: `#cloud-config
users:
  - name: custom
    groups: docker
    shell: /bin/zsh
packages:
  - nginx`,
		},
		{
			name: "CustomConfigWithUsersObject",
			customCloudConfig: `#cloud-config
users:
  name: custom
  groups: docker
  shell: /bin/zsh
packages:
  - nginx`,
		},
		{
			name: "CustomConfigWithUsersString",
			customCloudConfig: `#cloud-config
users: custom
packages:
  - nginx`,
		},
		{
			name:              "InvalidCustomCloudConfig",
			customCloudConfig: `invalid: yaml: content:`,
			expectError:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc, err := NewCloudConfig(tt.customCloudConfig)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, cc)
			require.NotNil(t, cc.config)
		})
	}
}

func TestCloudConfig_Render(t *testing.T) {
	tests := []struct {
		name           string
		setupConfig    func(*CloudConfig)
		expectedPrefix string
	}{
		{
			name: "EmptyConfig",
			setupConfig: func(cc *CloudConfig) {
				// No setup needed, empty config
			},
			expectedPrefix: "#cloud-config\n",
		},
		{
			name: "ConfigWithPackages",
			setupConfig: func(cc *CloudConfig) {
				cc.SetSection("packages", []string{"nginx", "curl"})
			},
			expectedPrefix: "#cloud-config\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := initEmptyCloudConfig()
			tt.setupConfig(cc)

			result, err := cc.Render()
			require.NoError(t, err)
			require.Contains(t, result, tt.expectedPrefix)
		})
	}
}

func TestCloudConfig_SetSection(t *testing.T) {
	tests := []struct {
		name       string
		sectionKey string
		value      any
	}{
		{
			name:       "SetString",
			sectionKey: "hostname",
			value:      "test-host",
		},
		{
			name:       "SetList",
			sectionKey: "packages",
			value:      []string{"nginx", "curl"},
		},
		{
			name:       "SetMap",
			sectionKey: "users",
			value:      map[string]any{"name": "packer", "shell": "/bin/bash"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := initEmptyCloudConfig()
			result := cc.SetSection(tt.sectionKey, tt.value)

			// Should return the same instance
			require.Equal(t, cc, result)

			// Value should be set
			require.Equal(t, tt.value, cc.config[tt.sectionKey])
		})
	}
}

func TestCloudConfig_AppendSection(t *testing.T) {
	tests := []struct {
		name        string
		sectionKey  string
		initialData any
		appendData  []any
		expected    any
	}{
		{
			name:        "AppendToNonExistentSection",
			sectionKey:  "packages",
			initialData: nil,
			appendData:  []any{"nginx"},
			expected:    []any{"nginx"},
		},
		{
			name:        "AppendToArray",
			sectionKey:  "packages",
			initialData: []any{"nginx"},
			appendData:  []any{"curl", "git"},
			expected:    []any{"nginx", "curl", "git"},
		},
		{
			name:        "AppendSingleValueToExistingValue",
			sectionKey:  "users",
			initialData: map[string]any{"name": "existing"},
			appendData:  []any{map[string]any{"name": "new"}},
			expected:    []any{map[string]any{"name": "existing"}, map[string]any{"name": "new"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := initEmptyCloudConfig()

			// Set initial data if provided
			if tt.initialData != nil {
				cc.config[tt.sectionKey] = tt.initialData
			}

			// Append data
			result := cc.AppendSection(tt.sectionKey, tt.appendData...)

			// Should return the same instance
			require.Equal(t, cc, result)

			// Check expected result
			require.Equal(t, tt.expected, cc.config[tt.sectionKey])
		})
	}
}

func TestCloudConfig_AppendUserForSSH(t *testing.T) {
	tests := []struct {
		name         string
		sshUsername  string
		sshPublicKey string
	}{
		{
			name:         "ValidUserAndKey",
			sshUsername:  "packer",
			sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
		},
		{
			name:         "EmptyUsername",
			sshUsername:  "",
			sshPublicKey: "ssh-rsa AAAAB3NzaC1yc2E...",
		},
		{
			name:         "EmptyKey",
			sshUsername:  "packer",
			sshPublicKey: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := initEmptyCloudConfig()
			result := cc.AppendUserForSSH(tt.sshUsername, tt.sshPublicKey)

			// Should return the same instance
			require.Equal(t, cc, result)

			// Check that users section exists
			users, exists := cc.config["users"]
			require.True(t, exists)

			// Check that users is a slice
			usersSlice, ok := users.([]any)
			require.True(t, ok)
			require.Len(t, usersSlice, 1)

			// Check user properties
			user, ok := usersSlice[0].(map[string]any)
			require.True(t, ok)

			if tt.sshUsername != "" {
				require.Equal(t, tt.sshUsername, user["name"])
			}

			if tt.sshPublicKey != "" {
				keys, ok := user["ssh-authorized-keys"].([]string)
				require.True(t, ok)
				require.Contains(t, keys, tt.sshPublicKey)
			}
		})
	}
}

func TestCloudConfig_RenderGolden(t *testing.T) {
	tests := []struct {
		name              string
		sshUsername       string
		sshPublicKey      string
		customCloudConfig string
		expectNewError    bool
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
			expectNewError:    true,
		},
	}

	expectedDir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc, err := NewCloudConfig(tt.customCloudConfig)
			if tt.expectNewError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			cc.AppendUserForSSH(tt.sshUsername, tt.sshPublicKey)
			actual, err := cc.Render()
			require.NoError(t, err)
			expectedDir.String(t, tt.name+".yaml", actual)
		})
	}
}
