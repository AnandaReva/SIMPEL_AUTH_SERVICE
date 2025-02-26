package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// GenerateHMAC membuat HMAC-SHA256 dari teks menggunakan kunci tertentu
func GenerateHMAC(text string, key string) (string, error) {
	if text == "" || key == "" {
		return "", errors.New("missing text or key")
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil)), nil // Perbaikan return value (hapus koma)
}

// GeneratePBKDF2 menggunakan PBKDF2 untuk hashing password
func GeneratePBKDF2(text string, salt string, length int, iterations int) (string, error) {
	if text == "" || salt == "" {
		return "", errors.New("missing text or salt")
	}
	if length <= 0 {
		return "", errors.New("length must be greater than 0")
	}
	if iterations <= 0 {
		return "", errors.New("iterations must be greater than 0")
	}

	// Menggunakan SHA256 sebagai hash function untuk PBKDF2
	hash := pbkdf2.Key([]byte(text), []byte(salt), iterations, length, sha256.New)

	return hex.EncodeToString(hash), nil // Mengembalikan hasil hashing dalam format hex
}

// XOR encoding menggunakan karakter dari signature sebagai key
func XorEncode(input string, key byte) (string, error) {
	if input == "" {
		return "", errors.New("input must not be empty")

	}
	if key == 0 {
		return "", errors.New("key must not be 0")
	}

	var encoded strings.Builder
	for i := 0; i < len(input); i++ {
		encoded.WriteByte(input[i] ^ key)
	}
	return encoded.String(), nil
}
