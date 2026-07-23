// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

// Package testutil provides helper functions for testing.
package testutil

import (
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/stretchr/testify/require"
)

const (
	ErrorKey         = "error"
	GeneratedDataKey = "generated_data"
)

func RequireGeneratedDataGet(t *testing.T, state multistep.StateBag, key string, expected any) {
	genDataResult := state.Get(GeneratedDataKey)
	require.NotNil(t, genDataResult, "Expected generated_data to be stored in state")

	genDataMap, ok := genDataResult.(map[string]any)
	require.True(t, ok, "Expected generated_data to be of type map[string]any, got %T", genDataResult)

	actual, ok := genDataMap[key]
	require.True(t, ok, "Expected %q to be stored in generated data", key)
	require.Equal(t, expected, actual)
}

func RequireStateGets(t *testing.T, state multistep.StateBag, kv map[string]any) {
	for key, expected := range kv {
		actual, ok := state.GetOk(key)
		require.True(t, ok, "Expected %q to be stored in state", key)
		require.Equal(t, expected, actual)
	}
}

func RequireStateGet(t *testing.T, state multistep.StateBag, key string, expected any) {
	actual, ok := state.GetOk(key)
	require.True(t, ok, "Expected %q to be stored in state", key)
	require.Equal(t, expected, actual)
}

func RequireStateNotSet(t *testing.T, state multistep.StateBag, key string) {
	_, ok := state.GetOk(key)
	require.False(t, ok, "Expected %q not to be stored in state", key)
}

func RequireActionContinue(t *testing.T, state multistep.StateBag, action multistep.StepAction) {
	require.Equal(t, multistep.ActionContinue, action, "Expected action to be ActionContinue, error: %v", state.Get(ErrorKey))
}

func RequireActionHalt(t *testing.T, state multistep.StateBag, action multistep.StepAction) {
	require.Equal(t, multistep.ActionHalt, action, "Expected action to be ActionHalt")
	require.NotNil(t, state.Get(ErrorKey), "Expected error to be stored in state")
}
