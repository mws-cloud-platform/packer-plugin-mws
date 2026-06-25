package mwsexport

import (
	"context"
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
)

type stepPrepareS3Keys struct {
	S3Config
}

func (s *stepPrepareS3Keys) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	driver := state.Get(mws.DriverKey).(mws.Driver)
	ui := state.Get(mws.UIKey).(packer.Ui)
	prefix := state.Get(mws.UUIDPrefixKey).(string)

	accessKey := s.AccessKey
	secretKey := s.SecretKey

	if s.ServiceAccount != "" {
		ui.Sayf("Creating HMAC Key...")
		var err error
		accessKey, secretKey, err = driver.CreateHMACKey(ctx, mws.CreateHMACKeyParams{
			ServiceAccout: s.ServiceAccount,
			HMACKeyName:   prefix + "tmp-hmac-key",
		})
		if err != nil {
			mws.ActionHaltWithError(state, fmt.Errorf("create hmac key: %w", err))
		}
		ui.Sayf("HMAC Key Created")
	}

	if accessKey == "" {
		mws.ActionHaltWithError(state, consterr.Error("access_key is not set"))
	}
	if secretKey == "" {
		mws.ActionHaltWithError(state, consterr.Error("secret_key is not set"))
	}

	state.Put(mws.S3AccessKeyKey, accessKey)
	state.Put(mws.S3SecretKeyKey, secretKey)

	return multistep.ActionContinue
}

func (s *stepPrepareS3Keys) Cleanup(state multistep.StateBag) {
	// TODO delete keys if created by packer
}
