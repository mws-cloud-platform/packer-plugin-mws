// Copyright IBM Corp. 2020, 2025
// SPDX-License-Identifier: MPL-2.0

package mws

import (
	_ "embed"
	"fmt"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
)

//go:embed test-fixtures/template.pkr.hcl
var testBuilderHCL2Basic string

func TestAccMWSBuilder(t *testing.T) {
	testCase := &acctest.PluginTestCase{
		Name:     "mws_builder_basic_test",
		Template: testBuilderHCL2Basic,
		Type:     "mws",
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}
