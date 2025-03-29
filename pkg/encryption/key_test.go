package encryption_test

import (
	"testing"

	"github.com/llravell/go-pass/pkg/encryption"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const masterPassword = "secret pass"

func TestKey(t *testing.T) {
	t.Run("keys generated from a single pass are equal", func(t *testing.T) {
		key1 := encryption.GenerateKeyFromMasterPass(masterPassword)
		key2 := encryption.GenerateKeyFromMasterPass(masterPassword)

		assert.Equal(t, key1.String(), key2.String())
	})

	t.Run("keys generated from different passwords are not equal", func(t *testing.T) {
		key1 := encryption.GenerateKeyFromMasterPass(masterPassword)
		key2 := encryption.GenerateKeyFromMasterPass(masterPassword + "sss")

		assert.NotEqual(t, key1.String(), key2.String())
	})

	t.Run("encrypt and decrypt", func(t *testing.T) {
		text := "some plain text"
		key := encryption.GenerateKeyFromMasterPass(masterPassword)

		ciphertext, err := key.Encrypt(text)
		require.NoError(t, err)

		decrypted, err := key.Decrypt(ciphertext)
		require.NoError(t, err)

		assert.Equal(t, text, decrypted)
	})
}
