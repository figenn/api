package utils

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}

	hashedPassword := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(append(salt, hashedPassword...)), nil
}

func ComparePassword(hashedPassword, password string) bool {
	decoded, err := base64.RawStdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false
	}

	salt := decoded[:16]
	storedHash := decoded[16:]

	computedHash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)

	return compareByteSlices(storedHash, computedHash)
}

func compareByteSlices(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
