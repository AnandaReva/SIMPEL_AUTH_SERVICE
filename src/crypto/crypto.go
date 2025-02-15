package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

func GenerateHMAC(text string, key string) (string, string) {
	if text == "" || key == "" {
		return "", "Missing Text or Key"
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil)), ""
}



// GeneratePBKDF2 replaces Argon2 for password hashing
func GeneratePBKDF2(text string, salt string, length int, iterations int) (string, string) {
	if text == "" || salt == "" {
		return "", "Missing text or salt"
	}

	// Menggunakan SHA256 sebagai hash function untuk PBKDF2
	hash := pbkdf2.Key([]byte(text), []byte(salt), iterations, length, sha256.New)

	// Debug: Cetak hasil dalam bentuk hex untuk memastikan sesuai dengan frontend
	hashHex := hex.EncodeToString(hash)

	fmt.Println("Text (password):", text)
	fmt.Println("Salt:", salt)
	fmt.Println("Iterations:", iterations)
	fmt.Println("Derived Key (Hex):", hashHex)

	return hashHex, ""
}
