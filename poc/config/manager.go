/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"encoding/json"
	"errors"
	"path"
)

// Manager: config manager. Manages configs.
type Manager struct {
	File FileProvider
}

// given a URI to a location where a consortium folder and settings should be created, do so
// Note: this is a testing function
func (m *Manager) CreateConsortium(uri string) error {

	err := m.File.MkDir(path.Join(uri, "consortium"))
	if err != nil {
		return errors.New("cannot create consortium at " + uri + ". " + err.Error())
	}

	domain, _ := path.Split(uri)

	consconf := ConsortiumConfig{
		Domain:       domain,
		Policy:       ConsortiumPolicy{},
		Stakeholders: nil,
		Signatures:   nil,
		Previous:     "",
	}

	data, err := json.Marshal(consconf)
	if err != nil {
		return err
	}

	println(data)

	return nil
}
