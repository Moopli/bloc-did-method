/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

// FileProvider is an interface for saving and loading files named using the hashes of their contents
type FileProvider interface {
	// Open opens and verifies a file, returning its contents
	Open(filePath string) ([]byte, error)
	// Save saves the given data as a file to the given location, with the given file extension, with the name being a hash of the contents
	Save(loc, ext string, data []byte) (string, error)
	// MkDir makes a directory with the given name
	MkDir(dir string) error
}
