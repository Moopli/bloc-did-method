/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"encoding/json"
	"errors"
	"os"
	"path"
)

// Manager: config manager. Manages configs.
type Manager struct {
	HashFile FileProvider
}

// TODO: redesign Manager to be based around building and applying ChangeSets instead of composing separate file operations

// SaveFile saves data to a file named with the SHA-256 hash of the file contents, within the directory dir
// Returns the name of the file
func SaveFile(filePath string, data []byte) (string, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}

	defer file.Close()

	_, err = file.Write(data)

	return filePath, nil
}

// given a URI to a location where a consortium folder and settings should be created, do so
// Note: this is a testing function
func (m *Manager) CreateConsortium(uri string) (*ConsortiumConfig, error) {
	err := os.Mkdir(path.Join(uri, "consortium"), 0774)
	if err != nil {
		return nil, errors.New("cannot create consortium at " + uri + ". " + err.Error())
	}

	consconf := ConsortiumConfig{
		Domain:       uri,
		Policy:       ConsortiumPolicy{},
		Stakeholders: nil,
		Signatures:   nil,
		Previous:     "",
	}

	data, err := json.Marshal(consconf)
	if err != nil {
		return nil, err
	}

	println(data)

	fp, err := SaveFile(path.Join(uri, "consortium", "conf.json"), data)
	if err != nil {
		return nil, err
	}

	println(fp)

	return &consconf, nil
}
