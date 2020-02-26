/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package hashlink

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"strings"

	multibase "github.com/multiformats/go-multibase"
	multihash "github.com/multiformats/go-multihash"
)

var ErrFileInvalid = errors.New("file hash invalid")

// SaveFile saves data to a file named with the SHA-256 hash of the file contents, within the directory dir
// Returns the name of the file
func SaveFile(dir, ext string, data []byte) (string, error) {

	hash, err := multihash.Sum([]byte(data), multihash.SHA2_256, -1)
	if err != nil {
		return "", err
	}

	name, err := multibase.Encode(multibase.Base58BTC, hash)
	if err != nil {
		return "", err
	}

	//println("data:", string(data))
	//println("data hash:", name)

	if ext[0] != '.' {
		ext = "." + ext
	}

	filePath := path.Join(dir, name+ext)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	_, err = file.Write(data)

	return filePath, nil
}

// LoadVerify loads a file, and verifies that the hash in the name is the SHA256 base58BTC hash of the contents.
func LoadVerify(filePath string) ([]byte, error) {
	//	 split the fileName to get the hash
	_, f := path.Split(filePath)

	fileName := strings.Split(f, ".")[0]

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	//println("contents:", string(data))

	hash, err := multihash.Sum(data, multihash.SHA2_256, -1)
	if err != nil {
		return nil, err
	}

	hashString, err := multibase.Encode(multibase.Base58BTC, hash)
	if err != nil {
		return nil, err
	}

	//println("file hash:", hashString)
	//println("file name:", fileName)

	if hashString != fileName {
		return nil, ErrFileInvalid
	}

	return data, nil
}
