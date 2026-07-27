// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"context"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -destination=mock/step_create_hmac_key_driver_mock.go . StepCreateHMACKeyDriver

type StepCreateHMACKeyDriver interface {
	CreateHMACKey(ctx context.Context, serviceAccount, name string) (accessKey string, secretKey string, err error)
	DeleteHMACKey(ctx context.Context, serviceAccount, name string) error
}
