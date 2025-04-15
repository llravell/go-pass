package encryption_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/llravell/go-pass/pkg/encryption"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const masterPassword = "secret pass"

func TestEncryption(t *testing.T) {
	t.Run("encrypt and decrypt", func(t *testing.T) {
		data := []byte("some plain text")
		key := encryption.GenerateKeyFromMasterPass(masterPassword)

		cipherdata, err := encryption.Encrypt(key, data)
		require.NoError(t, err)

		decrypted, err := encryption.Decrypt(key, cipherdata)
		require.NoError(t, err)

		assert.Equal(t, data, decrypted)
	})

	t.Run("encrypt and decrypt string", func(t *testing.T) {
		text := "some plain text"
		key := encryption.GenerateKeyFromMasterPass(masterPassword)

		cipherdata, err := encryption.EncryptString(key, text)
		require.NoError(t, err)

		decrypted, err := encryption.DecryptString(key, cipherdata)
		require.NoError(t, err)

		assert.Equal(t, text, decrypted)
	})

	t.Run("encrypt with reader and decrypt with writer", func(t *testing.T) {
		text := "some plain text"
		resultBuf := &bytes.Buffer{}
		key := encryption.GenerateKeyFromMasterPass(masterPassword)

		reader, err := encryption.NewEncryptReader(key, strings.NewReader(text))
		writer := encryption.NewDecryptWriter(key, resultBuf)
		require.NoError(t, err)

		encrypted, err := io.ReadAll(reader)
		require.NoError(t, err)

		_, err = writer.Write(encrypted)
		require.NoError(t, err)

		assert.Equal(t, text, resultBuf.String())
	})
}
