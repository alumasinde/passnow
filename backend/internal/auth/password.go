package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(plain string, cost int) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// VerifyPassword returns true only on an exact match. Callers must treat
// "wrong password" and "user not found" identically upstream (constant
// generic error) to avoid user-enumeration via response differences.
func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
