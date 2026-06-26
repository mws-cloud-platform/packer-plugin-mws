// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

func NewCloudConfig(customCloudConfig string) (*CloudConfig, error) {
	c := initEmptyCloudConfig()
	if customCloudConfig != "" {
		if err := yaml.Unmarshal([]byte(customCloudConfig), c.config); err != nil {
			return nil, fmt.Errorf("unmarshal cloud-config: %w", err)
		}
	}

	return c, nil
}

type CloudConfig struct {
	config map[string]any
}

func (c *CloudConfig) Render() (string, error) {
	if c == nil {
		c = initEmptyCloudConfig()
	}
	result, err := yaml.Marshal(c.config)
	if err != nil {
		return "", fmt.Errorf("marshal cloud-config: %w", err)
	}
	return "#cloud-config\n" + string(result), nil
}

func (c *CloudConfig) SetSection(sectionName string, value any) *CloudConfig {
	if c == nil {
		c = initEmptyCloudConfig()
	}
	c.config[sectionName] = value
	return c
}

func (c *CloudConfig) AppendSection(sectionName string, values ...any) *CloudConfig {
	if c == nil {
		c = initEmptyCloudConfig()
	}
	if section, ok := c.config[sectionName]; ok {
		if v, ok := section.([]any); ok {
			values = append(v, values...)
		} else {
			values = append([]any{section}, values...)
		}
	}
	c.config[sectionName] = values
	return c
}

func (c *CloudConfig) AppendUserForSSH(sshUsername, sshPublicKey string) *CloudConfig {
	user := map[string]any{
		"name":                sshUsername,
		"groups":              "sudo",
		"shell":               "/bin/bash",
		"sudo":                "ALL=(ALL) NOPASSWD:ALL",
		"ssh-authorized-keys": []string{sshPublicKey},
	}
	return c.AppendSection("users", user)
}

func initEmptyCloudConfig() *CloudConfig {
	return &CloudConfig{
		config: make(map[string]any),
	}
}
