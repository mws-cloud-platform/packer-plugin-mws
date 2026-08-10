// Copyright 2026 MTS Web Services, LLC.
// SPDX-License-Identifier: MPL-2.0

package steps

import (
	"io/fs"
	"os"
)

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -typed -destination=mock/file_writer_mock.go . FileWriter

type FileWriter interface {
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

var _ FileWriter = RealFileWriter{}

type RealFileWriter struct{}

func (fw RealFileWriter) WriteFile(name string, data []byte, perm fs.FileMode) error {
	return os.WriteFile(name, data, perm)
}
