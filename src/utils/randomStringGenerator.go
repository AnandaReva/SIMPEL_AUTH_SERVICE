package utils

import (
	"math/rand"
	"time"
)

func RandomStringGenerator(length int) (string, string) {
	if length <= 0 {
		return "", "length must be greater than 0" // Return an error string if the length is invalid
	}

	var charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano()) // Inisialisasi seed untuk menghasilkan angka acak

	result := make([]byte, length)

	for i := range result {
		result[i] = charSet[rand.Intn(len(charSet))] // Pilih karakter acak dari charSet
	}

	return string(result), "" // Return the generated string and an empty string indicating no error
}
