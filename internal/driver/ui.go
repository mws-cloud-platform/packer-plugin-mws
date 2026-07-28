// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"context"

	"github.com/hashicorp/packer-plugin-sdk/packer"
	mwserrors "go.mws.cloud/go-sdk/mws/errors"
)

func DeleteWithUI(
	ctx context.Context,
	ui packer.Ui,
	resourceType string,
	resourceName string,
	deleteMethod func(context.Context, string) error,
) {
	ui.Sayf("Deleting %s...", resourceType)
	err := deleteMethod(ctx, resourceName)
	switch {
	case err == nil:
		ui.Sayf("%s %q deleted", resourceType, resourceName)
	case mwserrors.IsAPIErrorNotFoundStatus(err):
		ui.Sayf("%s %q not found", resourceType, resourceName)
	default:
		ui.Errorf("Error deleting %s %q. Please delete it manually.\n"+
			"Error: %v.", resourceType, resourceName, err)
	}
}

func DeleteSubWithUI(
	ctx context.Context,
	ui packer.Ui,
	resourceType string,
	resourceName string,
	parentName string,
	deleteMethod func(context.Context, string, string) error,
) {
	ui.Sayf("Deleting %s...", resourceType)
	err := deleteMethod(ctx, parentName, resourceName)
	switch {
	case err == nil:
		ui.Sayf("%s %q from %q deleted", resourceType, resourceName, parentName)
	case mwserrors.IsAPIErrorNotFoundStatus(err):
		ui.Sayf("%s %q not found in %q", resourceType, resourceName, parentName)
	default:
		ui.Errorf("Error deleting %s %q from %q. Please delete it manually.\n"+
			"Error: %v.", resourceType, resourceName, parentName, err)
	}
}
