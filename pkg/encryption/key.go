package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

var ErrShortCiphertext = errors.New("ciphertext too short")

type Key struct {
	hash []byte
}

func GenerateKeyFromMasterPass(masterPassword string) *Key {
	hash := sha256.Sum256([]byte(masterPassword))

	return &Key{
		hash: hash[:],
	}
}

func (key *Key) String() string {
	return base64.StdEncoding.EncodeToString(key.hash)
}

func (key *Key) Encrypt(text string) (string, error) {
	block, err := aes.NewCipher(key.hash)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (key *Key) Decrypt(ciphertext string) (string, error) {
	cipherData, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key.hash)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherData) < nonceSize {
		return "", ErrShortCiphertext
	}

	nonce, cipherData := cipherData[:nonceSize], cipherData[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
