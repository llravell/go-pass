package encryption_test

import (
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
}
