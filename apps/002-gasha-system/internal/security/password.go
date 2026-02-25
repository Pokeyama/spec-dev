package security

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func HashPassword(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func ComparePassword(hash, raw string) bool {
	rawHash := HashPassword(raw)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(rawHash)) == 1
}
