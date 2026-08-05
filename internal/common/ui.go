// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"unicode"

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
	ui.Sayf("Deleting %s %q...", resourceType, resourceName)
	err := deleteMethod(ctx, resourceName)
	switch {
	case err == nil:
		ui.Sayf("%s %q deleted", UpperFirst(resourceType), resourceName)
	case mwserrors.IsAPIErrorNotFoundStatus(err):
		ui.Sayf("%s %q not found", UpperFirst(resourceType), resourceName)
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
	parentType string,
	parentName string,
	deleteMethod func(context.Context, string, string) error,
) {
	ui.Sayf("Deleting %s %q from %s %q...", resourceType, resourceName, parentType, parentName)
	err := deleteMethod(ctx, parentName, resourceName)
	switch {
	case err == nil:
		ui.Sayf("%s %q from %s %q deleted", UpperFirst(resourceType), resourceName, parentType, parentName)
	case mwserrors.IsAPIErrorNotFoundStatus(err):
		ui.Sayf("%s %q not found in %s %q", UpperFirst(resourceType), resourceName, parentType, parentName)
	default:
		ui.Errorf("Error deleting %s %q from %s %q. Please delete it manually.\n"+
			"Error: %v.", resourceType, resourceName, parentType, parentName, err)
	}
}

func UpperFirst(s string) string {
	if len(s) == 0 {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
