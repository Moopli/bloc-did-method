/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package hashlink

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Hashlink(t *testing.T) {
	err := os.Mkdir("tmp", 0774)
	if err != nil {
		require.Contains(t, err.Error(), "file exists")
	}

	t.Run("success: write then read and verify", func(t *testing.T) {
		filePath, err := SaveFile("tmp/", ".dat", []byte("test data"))
		require.NoError(t, err)

		data, err := LoadVerify(filePath)
		require.NoError(t, err)

		require.Equal(t, []byte("test data"), data)
	})

	t.Run("fail: write, rename file, read", func(t *testing.T) {
		filePath, err := SaveFile("tmp/", ".dat", []byte("test data 2"))
		require.NoError(t, err)

		newPath, _ := path.Split(filePath)
		newPath = path.Join(newPath, "badfile.oops")

		err = os.Rename(filePath, newPath)
		require.NoError(t, err)

		_, err = LoadVerify(newPath)
		require.Error(t, err)
		require.Equal(t, ErrFileInvalid, err)
	})

	t.Run("fail: write, change file, read", func(t *testing.T) {
		filePath, err := SaveFile("tmp/", ".dat", []byte("lorem ipsum dolor"))
		require.NoError(t, err)

		err = ioutil.WriteFile(filePath, []byte("Lorem ipsum dolor"), 0644)

		_, err = LoadVerify(filePath)
		require.Error(t, err)
		require.Equal(t, ErrFileInvalid, err)
	})

}
