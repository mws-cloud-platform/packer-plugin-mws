// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package example

import (
	_ "embed"
)

//go:embed builder/build.pkr.hcl
var BuilderHCL string

//go:embed export/build.pkr.hcl
var ExportHCL string
