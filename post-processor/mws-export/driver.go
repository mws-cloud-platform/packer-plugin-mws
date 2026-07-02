// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import (
	"context"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -package=mock -destination=mock/driver_mock.go . Driver

type Driver interface {
	CreateHMACKey(ctx context.Context, serviceAccount, name string) (accessKey string, secretKey string, err error)
	DeleteHMACKey(ctx context.Context, serviceAccount, name string) error
}
