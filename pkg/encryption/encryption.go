package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var ErrShortCiphertext = errors.New("ciphertext too short")

func Encrypt(key *Key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key.hash)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func EncryptString(key *Key, text string) (string, error) {
	encrypted, err := Encrypt(key, []byte(text))
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func Decrypt(key *Key, cipherdata []byte) ([]byte, error) {
	block, err := aes.NewCipher(key.hash)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(cipherdata) < nonceSize {
		return nil, ErrShortCiphertext
	}

	nonce, data := cipherdata[:nonceSize], cipherdata[nonceSize:]

	decrypted, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}

func DecryptString(key *Key, ciphertext string) (string, error) {
	cipherdata, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	data, err := Decrypt(key, cipherdata)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type EncryptReader struct {
	nonce  []byte
	stream cipher.Stream
	reader io.Reader
}

func NewEncryptReader(key *Key, reader io.Reader) (*EncryptReader, error) {
	block, err := aes.NewCipher(key.Hash())
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return &EncryptReader{
		nonce:  nonce,
		reader: reader,
		stream: cipher.NewCTR(block, nonce),
	}, nil
}

func (r *EncryptReader) Read(p []byte) (int, error) {
	if len(r.nonce) > 0 {
		n := copy(p, r.nonce)
		r.nonce = r.nonce[n:]

		return n, nil
	}

	n, err := r.reader.Read(p)
	if n > 0 {
		r.stream.XORKeyStream(p[:n], p[:n])
	}

	return n, err
}

type DecryptWriter struct {
	key    *Key
	stream cipher.Stream
	writer io.Writer
}

func NewDecryptWriter(key *Key, writer io.Writer) *DecryptWriter {
	return &DecryptWriter{
		key:    key,
		writer: writer,
		stream: nil,
	}
}

func (w *DecryptWriter) Write(p []byte) (int, error) {
	if w.stream == nil {
		nonce, data := p[:aes.BlockSize], p[aes.BlockSize:]

		block, err := aes.NewCipher(w.key.Hash())
		if err != nil {
			return 0, err
		}

		w.stream = cipher.NewCTR(block, nonce)
		p = data
	}

	w.stream.XORKeyStream(p, p)

	return w.writer.Write(p)
}
