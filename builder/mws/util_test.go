// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mws_test

import (
	"bytes"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"github.com/stretchr/testify/require"
)

const (
	packerPrefix            = "packer-"
	testProjectName         = "test-project"
	testDiskName            = "test-disk"
	testExternalAddressName = "test-external-address"
	testNetworkName         = "test-network"
	testSubnetName          = "test-subnet"
	testVirtualMachineName  = "test-vm"
	testImageName           = "test-image"
	testImageDescription    = "Test image description"
	testInternalAddress     = "192.168.0.10"
	testExternalAddress     = "10.20.30.40"
	testSSHPublicKey        = "test-public-key"
	testSourceImage         = "test-source-image"

	defaultDiskName            = packerPrefix + "disk"
	defaultExternalAddressName = packerPrefix + "external-address"
	defaultNetworkName         = packerPrefix + "network"
	defaultSubnetName          = packerPrefix + "subnet"
	defaultVirtualMachineName  = packerPrefix + "vm"
	defaultImageName           = packerPrefix + "image"
)

func requireGeneratedDataGet(t *testing.T, state multistep.StateBag, key string, expected any) {
	genDataResult := state.Get(mws.GeneratedDataKey)
	require.NotNil(t, genDataResult, "Expected generated_data to be stored in state")

	genDataMap, ok := genDataResult.(map[string]any)
	require.True(t, ok, "Expected generated_data to be of type map[string]any, got %T", genDataResult)

	actual, ok := genDataMap[key]
	require.True(t, ok, "Expected %q to be stored in generated data", key)
	require.Equal(t, expected, actual)
}

func requireStateGet(t *testing.T, state multistep.StateBag, key string, expected any) {
	actual, ok := state.GetOk(key)
	require.True(t, ok, "Expected %q to be stored in state", key)
	require.Equal(t, expected, actual)
}

func requireActionContinue(t *testing.T, state multistep.StateBag, action multistep.StepAction) {
	require.Equal(t, multistep.ActionContinue, action, "Expected action to be ActionContinue, error: %v", state.Get(mws.ErrorKey))
}

func requireActionHalt(t *testing.T, state multistep.StateBag, action multistep.StepAction) {
	require.Equal(t, multistep.ActionHalt, action, "Expected action to be ActionHalt")
	require.NotNil(t, state.Get(mws.ErrorKey), "Expected error to be stored in state")
}

func prepareState(t *testing.T, config *mws.Config, driver mws.Driver) (*bytes.Buffer, multistep.StateBag) {
	state := new(multistep.BasicStateBag)

	config.SetDefaults()
	require.NoError(t, config.Validate())
	config.Communicator.SSHPublicKey = []byte(testSSHPublicKey)
	state.Put(mws.ConfigKey, config)
	state.Put(mws.DriverKey, driver)
	state.Put(mws.UuidPrefixKey, packerPrefix)
	writer := new(bytes.Buffer)
	ui := &packer.BasicUi{
		Writer: writer,
	}
	state.Put(mws.UiKey, ui)

	return writer, state
}
