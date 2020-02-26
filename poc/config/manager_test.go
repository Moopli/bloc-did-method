/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package config

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManager_CreateConsortium(t *testing.T) {
	err := os.Mkdir("tmp", 0774)
	if err != nil {
		require.True(t, errors.Is(err, os.ErrExist))
	}

	m := Manager{HashFile: &LocalFileProvider{}}
	_, err = m.CreateConsortium("tmp/")
	require.NoError(t, err)
}
