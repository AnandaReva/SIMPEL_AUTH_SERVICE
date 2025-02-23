package utils

import (
	"errors"
	"math/rand"
	"strconv"
	"time"
)

// Fungsi untuk menghasilkan string acak
func RandomStringGenerator(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than 0") // Mengembalikan error jika panjang tidak valid
	}

	var charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano()) // Inisialisasi seed untuk hasil yang lebih acak

	result := make([]byte, length)
	for i := range result {
		result[i] = charSet[rand.Intn(len(charSet))] // Pilih karakter acak dari charSet
	}

	return string(result), nil // Mengembalikan string hasil dan error nil jika tidak ada error
}

// Fungsi untuk menghasilkan angka acak dengan panjang tertentu (mengembalikan int)
func RandoNnumberGenerator(length int) (int, error) {
	if length <= 0 {
		return 0, errors.New("length must be greater than 0")
	}

	var charSet = "0123456789"
	rand.Seed(time.Now().UnixNano())

	result := make([]byte, length)
	for i := range result {
		result[i] = charSet[rand.Intn(len(charSet))] // Pilih angka acak dari charSet
	}

	number, err := strconv.Atoi(string(result)) // Konversi string ke integer
	if err != nil {
		return 0, err
	}

	return number, nil // Mengembalikan hasil dalam bentuk integer
}
