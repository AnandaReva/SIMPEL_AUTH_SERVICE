package handlers

import (
	"auth_service/db"
	"auth_service/logger"
	"auth_service/utils"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// MakeSalt menghasilkan salt acak sepanjang 16 karakter
func MakeSalt() (string, string) {
	salt, err := utils.RandomStringGenerator(16)
	if err != "" {
		return "", "Failed to create salt"
	}
	return salt, ""
}

// GenerateSaltedPassword membuat hash password dengan salt menggunakan HMAC-SHA256
func GenerateSaltedPassword(password string, salt string) (string, string) {
	if password == "" || salt == "" {
		return "", "Missing Password or Salt"
	}

	key := []byte(salt)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil)), ""
}

// GenerateHMAC membuat HMAC-SHA256 dari teks menggunakan kunci tertentu
/* func GenerateHMAC(text string, key string) (string, string) {
	if text == "" || key == "" {
		return "", "Missing Text or Key"
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil)), ""
} */

/*
{
    "username" : "master",
    "full_name" : "Master User",
    "password" : "master123"
}
*/

func Register(w http.ResponseWriter, r *http.Request) {
	// !!!NOTE : JANGAN MEMBERIKAN PESAN ERROR TERLALU DETAIL KE CLIENT

	var ctxKey HTTPContextKey = "requestID"
	referenceID, ok := r.Context().Value(ctxKey).(string)
	if !ok {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - Register - Execution completed in ", duration)
	}()

	// Inisialisasi result
	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	// porses bocy request
	param, _ := utils.Request(r)

	// Validasi input
	username, ok := param["username"].(string)
	if !ok || username == "" {
		logger.Error(referenceID, "ERROR - Register - Missing username")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	fullName, ok := param["full_name"].(string)
	if !ok || fullName == "" {
		logger.Error(referenceID, "ERROR - Register - Missing full_name")
		result.ErrorCode = "400002"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	password, ok := param["password"].(string)
	if !ok || password == "" {
		logger.Error(referenceID, "ERROR - Register - Missing password")
		result.ErrorCode = "400003"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	// Buat Salt
	salt, errSalt := MakeSalt()
	if errSalt != "" {
		logger.Error(referenceID, "ERROR - Register - Failed to generate salt: ", errSalt)
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	logger.Info(referenceID, "INFO - Register - Salt generated:", salt)

	// Buat Salted Password
	saltedPassword, errSaltedPass := GenerateSaltedPassword(password, salt)
	if errSaltedPass != "" {
		logger.Error(referenceID, "ERROR - Register - Failed to generate salted password: ", errSaltedPass)
		result.ErrorCode = "500001"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	logger.Info(referenceID, "INFO - Register - Salted password generated")

	// Ambil koneksi database
	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to get DB connection")
		result.ErrorCode = "500002"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	// Query untuk mengecek apakah username sudah ada
	queryCheckUsername := `SELECT COUNT(*) FROM sysuser.user WHERE username = $1;`

	var count int
	err = conn.Get(&count, queryCheckUsername, username)
	if err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to check existing username")
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	if count > 0 {
		logger.Error(referenceID, "ERROR - Register - Username already exists")
		result.ErrorCode = "409000"
		result.ErrorMessage = "Conflict"
		utils.Response(w, result)
		return
	}

	// Query untuk menyimpan user
	queryToRegister := `
	INSERT INTO sysuser.user (username, full_name, st, salt, saltedpassword, data, role) 
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id;
`

	var newUserId int
	err = conn.Get(&newUserId, queryToRegister, username, fullName, 1, salt, saltedPassword, "{}", "guest")
	if err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to insert new account")
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	logger.Info(referenceID, "INFO - Register - Successfully registered user with ID =", newUserId)

	// Sukses
	result.ErrorCode = "000000"
	result.ErrorMessage = ""
	result.Payload["status"] = "success"
	result.Payload["user_id"] = newUserId

	utils.Response(w, result)
}

// writeJSONResponse menulis response dalam format JSON
