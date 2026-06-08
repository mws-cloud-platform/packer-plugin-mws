package mws_test

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"github.com/stretchr/testify/require"
	computeref "go.mws.cloud/go-sdk/service/resources/references/compute"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
)

const (
	packerPrefix         = "packer-"
	testProjectName      = "test-project"
	testDiskName         = "test-disk"
	testExternalAddr     = "test-external-address"
	testNetworkName      = "test-network"
	testSubnetName       = "test-subnet"
	testVmName           = "test-vm"
	testImageName        = "test-image"
	testImageDescription = "Test image description"
	testSubnetCidr       = "192.168.0.0/16"
	testInternalAddress  = "192.168.0.10"
	testExternalAddress  = "10.20.30.40"
	testSSHPublicKey     = "test-public-key"
	testSourceImage      = "test-source-image"

	defaultDiskName            = packerPrefix + "disk"
	defaultExternalAddressName = packerPrefix + "external-address"
	defaultNetworkName         = packerPrefix + "network"
	defaultSubnetName          = packerPrefix + "subnet"
	defaultVmName              = packerPrefix + "vm"
	defaultImageName           = packerPrefix + "image"
)

var (
	testDiskRef = new(computeref.NewDiskRef(testProjectName, testDiskName))
)

func requireGeneratedDataGet(t *testing.T, state multistep.StateBag, key string, expected any) {
	genDataResult := state.Get(mws.GeneratedDataKey)
	require.NotNil(t, genDataResult, "Expected generated_data to be stored in state")

	genDataMap, ok := genDataResult.(map[string]any)
	require.True(t, ok, "Expected generated_data to be of type map[string]any, got %T", genDataResult)

	actual, ok := genDataMap[key]
	require.True(t, ok, "Expected `%s` to be stored in generated data", key)
	require.Equal(t, expected, actual)
}

func requireStateGet(t *testing.T, state multistep.StateBag, key string, expected any) {
	actual, ok := state.GetOk(key)
	require.True(t, ok, "Expected `%s` to be stored in state", key)
	require.Equal(t, expected, actual)
}

func requireActionContinue(t *testing.T, state multistep.StateBag, action multistep.StepAction) {
	require.Equal(t, multistep.ActionContinue, action, "Expected action to be ActionContinue, error: %v", state.Get(mws.ErrorKey))
}

func requireActionHalt(t *testing.T, state multistep.StateBag, action multistep.StepAction) {
	require.Equal(t, multistep.ActionHalt, action, "Expected action to be ActionHalt")
	require.NotNil(t, state.Get(mws.ErrorKey), "Expected error to be stored in state")
}

func requireOutput(t *testing.T, output string) {
	expectedDir := golden.NewDir(t)
	expectedDir.String(t, t.Name()+".out", output)
}
