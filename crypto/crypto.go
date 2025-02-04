package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateHMAC(text string, key string) (string, string) {
	if text == "" || key == "" {
		return "", "Missing Text or Key"
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil)), ""
}
