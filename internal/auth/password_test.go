package auth

import "testing"

func TestPasswordHashAndVerify(t *testing.T) {
	hash, err := HashPassword("correct horse battery staple", 4)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "correct horse battery staple" {
		t.Fatal("password hash must not equal plaintext")
	}
	if !VerifyPassword(hash, "correct horse battery staple") {
		t.Fatal("correct password was rejected")
	}
	if VerifyPassword(hash, "wrong password") {
		t.Fatal("wrong password was accepted")
	}
}

func TestVerifyPasswordRejectsMalformedHash(t *testing.T) {
	if VerifyPassword("not-a-bcrypt-hash", "password") {
		t.Fatal("malformed hash was accepted")
	}
}
