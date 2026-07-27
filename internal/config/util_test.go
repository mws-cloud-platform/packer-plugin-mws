// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package config_test

import (
	"errors"
	"testing"

	packerconfig "github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/stretchr/testify/require"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

type Config interface {
	SetDefaults()
	Validate() error
}

type ConfigTestCase struct {
	name    string
	raws    []any
	wantErr bool
}

func (tc *ConfigTestCase) ConfigTest(c Config, expectedDir *golden.Dir) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		err := packerconfig.Decode(c, nil, tc.raws...)
		c.SetDefaults()
		err = errors.Join(err, c.Validate())

		if tc.wantErr {
			require.Error(t, err)
			expectedDir.String(t, tc.name+".txt", err.Error())
		} else {
			require.NoError(t, err)
			expectedDir.JSON(t, tc.name+".json", c)
		}
	}
}
