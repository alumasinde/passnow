package tenantdb

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestCipherRoundTrip(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	c, err := NewCipher(key)
	if err != nil { t.Fatal(err) }
	encrypted, err := c.Encrypt("secret-password")
	if err != nil { t.Fatal(err) }
	if strings.Contains(encrypted, "secret-password") { t.Fatal("ciphertext contains plaintext") }
	plain, err := c.Decrypt(encrypted)
	if err != nil { t.Fatal(err) }
	if plain != "secret-password" { t.Fatalf("got %q", plain) }
}

func TestCipherRejectsBadKey(t *testing.T) {
	if _, err := NewCipher("bad"); err == nil { t.Fatal("expected error") }
}
