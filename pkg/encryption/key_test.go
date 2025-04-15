package encryption_test

import (
	"testing"

	"github.com/llravell/go-pass/pkg/encryption"
	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	t.Run("keys generated from a single pass are equal", func(t *testing.T) {
		key1 := encryption.GenerateKeyFromMasterPass("pass")
		key2 := encryption.GenerateKeyFromMasterPass("pass")

		assert.Equal(t, key1.String(), key2.String())
	})

	t.Run("keys generated from different passwords are not equal", func(t *testing.T) {
		key1 := encryption.GenerateKeyFromMasterPass("pass")
		key2 := encryption.GenerateKeyFromMasterPass("another_pass")

		assert.NotEqual(t, key1.String(), key2.String())
	})
}
