// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

//go:generate go run github.com/hashicorp/packer-plugin-sdk/cmd/packer-sdc@v0.6.9 struct-markdown

package config

import (
	"cmp"

	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

type ObjectStorageConfig struct {
	// MWS Cloud Platform Service Account used for generating temporal HMAC key
	// to access Object Storage. Required, unless `access_key` and `secret_key`
	// are provided.
	ServiceAccount string `mapstructure:"service_account" required:"false"`

	// HMAC key identifier for authenticating with Object Storage. Used if
	// `service_account` is not provided. Also requires `secret_key` to be
	// provided.
	AccessKey string `mapstructure:"access_key" required:"false"`
	// HMAC key secret for accessing Object Storage. Required if `access_key` is
	// provided.
	SecretKey string `mapstructure:"secret_key" required:"false"`

	// MWS Cloud Platform Object Storage endpoint (defaults to "https://storage.mwsapis.ru").
	ObjectStorageEndpoint string `mapstructure:"object_storage_endpoint" required:"false"`
	// MWS Cloud Platform Object Storage region where the bucket is located (defaults to "ru-central1").
	ObjectStorageRegion string `mapstructure:"object_storage_region" required:"false"`
}

func (c *ObjectStorageConfig) SetDefaults() {
	c.ObjectStorageEndpoint = cmp.Or(c.ObjectStorageEndpoint, DefaultObjectStorageEndpoint)
	c.ObjectStorageRegion = cmp.Or(c.ObjectStorageRegion, DefaultObjectStorageRegion)
}

func (c *ObjectStorageConfig) Validate() error {
	if (c.SecretKey == "" || c.AccessKey == "") && c.ServiceAccount == "" {
		return consterr.Error("Object Storage authentication is not provided, " +
			"provide service_account for hmac-key generation (recommended) or pair access_key, secret_key")
	}
	return nil
}
