// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package version

import "github.com/hashicorp/packer-plugin-sdk/version"

var (
	Version           = "0.5.0"
	VersionPrerelease = ""
	VersionMetadata   = ""
	PluginVersion     = version.NewPluginVersion(Version, VersionPrerelease, VersionMetadata)
)
