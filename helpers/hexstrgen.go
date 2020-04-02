package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

// HexStringGen generates a n-byte long hex string. Returns empty string if err is not nil.
// Note: 5 bytes takes 4 years to crack.
func HexStringGen(n int) (string, error) {

	b := make([]byte, n)
	
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	
	return hex.EncodeToString(b), nil
}
