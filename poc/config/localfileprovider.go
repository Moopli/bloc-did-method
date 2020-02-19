/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"github.com/trustbloc/bloc-did-method/poc/hashlink"
)

type LocalFileProvider struct{}

func (lfp *LocalFileProvider) Open(filePath string) ([]byte, error) {
	return hashlink.LoadVerify(filePath)
}

// Save saves the given data as a file to the given directory, with the given file extension, with the name being a hash of the contents
func (lfp *LocalFileProvider) Save(dir, ext string, data []byte) (string, error) {
	return hashlink.SaveFile(dir, ext, data)
}
