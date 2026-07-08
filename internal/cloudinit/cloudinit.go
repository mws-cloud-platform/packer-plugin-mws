// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package cloudinit

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func PrepareCloudConfig(sshUsername, sshPublicKey, customCloudConfig string) (string, error) {
	config := make(map[string]any)
	if customCloudConfig != "" {
		if err := yaml.Unmarshal([]byte(customCloudConfig), &config); err != nil {
			return "", fmt.Errorf("unmarshal custom cloud-config: %w", err)
		}
	}

	users := []any{map[string]any{
		"name":                sshUsername,
		"groups":              "sudo",
		"shell":               "/bin/bash",
		"sudo":                "ALL=(ALL) NOPASSWD:ALL",
		"ssh-authorized-keys": []string{sshPublicKey},
	}}

	if customUsers, ok := config["users"]; ok {
		if v, ok := customUsers.([]any); ok {
			users = append(users, v...)
		} else {
			users = append(users, customUsers)
		}
	}
	config["users"] = users

	result, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("marshal cloud-config: %w", err)
	}

	return "#cloud-config\n" + string(result), nil
}
