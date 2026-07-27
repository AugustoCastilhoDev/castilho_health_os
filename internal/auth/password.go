// Package auth holds password hashing and JWT issuance/verification —
// the primitives the auth service and Fiber middleware build on.
package auth

import "golang.org/x/crypto/bcrypt"

// bcryptCost is raised above bcrypt.DefaultCost (10) because this system
// handles LGPD-sensitive health data and the hashing only runs on login/
// registration, not on a hot path.
const bcryptCost = 12

func HashPassword(plain string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword reports whether plain matches hash, in constant time.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
