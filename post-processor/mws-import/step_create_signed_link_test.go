// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package mwsimport_test

import (
	"bytes"
	"context"
	"path"
	"testing"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/packer"
	mws "github.com/mws-cloud-platform/packer-plugin-mws/builder/mws"
	"github.com/mws-cloud-platform/packer-plugin-mws/internal/testutil"
	mwsimport "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-import"
	mockmws "github.com/mws-cloud-platform/packer-plugin-mws/post-processor/mws-import/mock"
	"go.mws.cloud/util-toolset/pkg/testing/golden"
	"go.mws.cloud/util-toolset/pkg/utils/consterr"
	"go.uber.org/mock/gomock"
)

var (
	testBucket       = "test-bucket"
	testKey          = "path/to/object.qcow2"
	testPath         = testBucket + "/" + testKey
	testPresignedURL = "https://test-bucket.storage.mwsapis.ru/path/to/object.qcow2?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Checksum-Mode=ENABLED&X-Amz-Credential=test&X-Amz-Date=test&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=test"
	errStepInternal  = consterr.Error("internal error")
)

func TestStepCreateSignedLink(t *testing.T) {
	t.Parallel()
	dir := golden.NewDir(t, golden.WithPath(path.Join("testdata", t.Name())), golden.WithRecreateOnUpdate())

	for _, tt := range []struct {
		name          string
		path          string
		prepare       func(multistep.StateBag, *mockmws.MockAWSClient)
		expectedError bool
	}{
		{
			name: "success",
			path: testPath,
			prepare: func(state multistep.StateBag, client *mockmws.MockAWSClient) {
				client.EXPECT().
					PresignGetObject(gomock.Any(), &s3.GetObjectInput{
						Bucket: &testBucket,
						Key:    &testKey,
					}, gomock.Any()).
					Return(&v4.PresignedHTTPRequest{
						URL: testPresignedURL,
					}, nil).
					Times(1)
			},
		},
		{
			name: "presign_get_object_error",
			path: testPath,
			prepare: func(state multistep.StateBag, client *mockmws.MockAWSClient) {
				client.EXPECT().
					PresignGetObject(gomock.Any(), &s3.GetObjectInput{
						Bucket: &testBucket,
						Key:    &testKey,
					}, gomock.Any()).
					Return(nil, errStepInternal).
					Times(1)
			},
			expectedError: true,
		},
		{
			name:          "malformed_path_missing_slash",
			path:          "bucket-without-key",
			expectedError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)
			client := mockmws.NewMockAWSClient(ctrl)

			state := new(multistep.BasicStateBag)
			state.Put(mwsimport.AWSClientKey, client)
			writer := new(bytes.Buffer)
			ui := &packer.BasicUi{Writer: writer}
			state.Put(mws.UIKey, ui)

			if tt.prepare != nil {
				tt.prepare(state, client)
			}

			step := &mwsimport.StepCreateSignedLink{
				Path: tt.path,
			}

			action := step.Run(context.Background(), state)
			if tt.expectedError {
				testutil.RequireActionHalt(t, state, action)
			} else {
				testutil.RequireActionContinue(t, state, action)
				testutil.RequireStateGet(t, state, mwsimport.ExternalURLKey, testPresignedURL)
			}
			dir.String(t, tt.name+".out", writer.String())
		})
	}
}
