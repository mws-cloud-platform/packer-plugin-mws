// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown

package config

import (
	"cmp"
	"os"

	"go.mws.cloud/go-sdk/mws"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

type AccessConfig struct {
	// The project identifier where resources will be created.
	// Can be specified using the `MWS_PROJECT` environment variable.
	Project string `mapstructure:"project" required:"false"`
	// The zone in which the VM will be created (defaults to "ru-central1-a")
	Zone string `mapstructure:"zone" required:"false"`

	// MWS Cloud Platform API base endpoint (defaults to "https://api.mwsapis.ru").
	// Can be specified using the `MWS_BASE_ENDPOINT` environment variable.
	BaseEndpoint string `mapstructure:"base_endpoint" required:"false"`
	// Path to the service account authorized key file used for authentication.
	// Has no effect if IAM token is set.
	// Can be specified using the `MWS_SERVICE_ACCOUNT_AUTHORIZED_KEY_PATH` environment variable.
	ServiceAccountAuthorizedKeyPath string `mapstructure:"service_account_authorized_key_path" required:"false"`
	// IAM token used for authentication.
	// Can be specified using the `MWS_TOKEN` environment variable.
	Token string `mapstructure:"token" required:"false"`
}

func (c *AccessConfig) SetDefaults() {
	c.Project = cmp.Or(c.Project, os.Getenv(mws.ProjectEnv))
	c.Zone = cmp.Or(c.Zone, DefaultZone)
}

func (c *AccessConfig) Validate() error {
	if c.Project == "" {
		return consterr.Error("project is not provided")
	}
	return nil
}
