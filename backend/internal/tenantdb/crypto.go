package tenantdb

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type Cipher struct{ aead cipher.AEAD }

func NewCipher(base64Key string) (*Cipher, error) {
	if base64Key == "" {
		return nil, fmt.Errorf("tenantdb: TENANT_DB_ENCRYPTION_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("tenantdb: encryption key must be base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("tenantdb: encryption key must decode to exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil { return nil, err }
	aead, err := cipher.NewGCM(block)
	if err != nil { return nil, err }
	return &Cipher{aead: aead}, nil
}

func (c *Cipher) Encrypt(plain string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil { return "", err }
	sealed := c.aead.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func (c *Cipher) Decrypt(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil { return "", err }
	n := c.aead.NonceSize()
	if len(raw) < n { return "", fmt.Errorf("tenantdb: invalid encrypted payload") }
	plain, err := c.aead.Open(nil, raw[:n], raw[n:], nil)
	if err != nil { return "", fmt.Errorf("tenantdb: decrypt credentials: %w", err) }
	return string(plain), nil
}
