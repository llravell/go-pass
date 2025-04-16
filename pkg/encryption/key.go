package encryption

import (
	"crypto/sha256"
	"encoding/base64"
)

type Key struct {
	hash []byte
}

func GenerateKeyFromMasterPass(masterPassword string) *Key {
	hash := sha256.Sum256([]byte(masterPassword))

	return &Key{
		hash: hash[:],
	}
}

func (key *Key) Hash() []byte {
	return key.hash
}

func (key *Key) String() string {
	return base64.StdEncoding.EncodeToString(key.hash)
}
