// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package driver

import "time"

type Config struct {
	Project                         string
	BaseEndpoint                    string
	ServiceAccountAuthorizedKeyPath string
	Token                           string
	CleanupTimeout                  time.Duration
}
