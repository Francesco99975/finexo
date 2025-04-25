package helpers

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateNonce() (string, error) {
	bytes := make([]byte, 16) // 16 bytes nonce
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
