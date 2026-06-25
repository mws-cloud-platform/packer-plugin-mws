// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsexport

import "fmt"

//nolint:revive // Very special constant for packer
const BuilderId = "packer.post-processor.mws-export"

type Artifact struct {
	path string
	url  string
}

//nolint:revive // Can not change packer interface
func (*Artifact) BuilderId() string {
	return BuilderId
}

//nolint:revive // Can not change packer interface
func (a *Artifact) Id() string {
	return a.url
}

func (a *Artifact) Files() []string {
	return []string{a.path}
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Exported artifact in: %s", a.path)
}

func (*Artifact) State(name string) any {
	return nil
}

func (a *Artifact) Destroy() error {
	return nil
}
