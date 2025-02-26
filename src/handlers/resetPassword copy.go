package handlers

import (
	"auth_service/configs"
	"auth_service/crypto"
	"auth_service/db"
	"auth_service/logger"
	"auth_service/mail"
	"auth_service/rds"
	"auth_service/utils"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// !NOTE : reset password procedure
/*	1. Client send request for reset password with email
2. Server check if email is exist in database , if exist, generate url with signature as param and send it via email
3. Client receive email and click the link, then user will be redirected to reset password page fill new password and send it back to server
4. Server check if signature is valid and not expired, then update password in database
*/

/* Modifikasi Rule Encoding Expiry
Ambil epoch time & konversi ke string
➜ Contoh: 1708890000
Ganti hanya digit pada indeks genap dengan huruf:
0 → k, 2 → b, 4 → d, 6 → f, 8 → h
Hasil: 1h0h8b9k0k0k
Sisipkan ke dalam urlSignature secara acak (misalnya setelah 5 karakter pertama).
*/

// Mapping angka ke huruf hanya untuk index genap
var numToChar = map[rune]string{
	'0': "k", '2': "b", '4': "d", '6': "f", '8': "h",
}

// Fungsi Encode & Inject Expiry dalam Signature
func encodeAndInjectExpiry(signature string, epoch int64) string {
	// Encode expiry
	strEpoch := strconv.FormatInt(epoch, 10)
	var encoded strings.Builder
	for i, ch := range strEpoch {
		if i%2 == 0 { // Ubah hanya indeks genap
			if mapped, exists := numToChar[ch]; exists {
				encoded.WriteString(mapped)
			} else {
				encoded.WriteRune(ch)
			}
		} else {
			encoded.WriteRune(ch)
		}
	}

	encodedExpiry := encoded.String()

	// Inject expiry ke signature di posisi tetap setelah karakter ke-5
	insertPos := 5
	return signature[:insertPos] + encodedExpiry + signature[insertPos:]
}

func Reset_Password(w http.ResponseWriter, r *http.Request) {
	var ctxKey HTTPContextKey = "requestID"
	referenceID, _ := r.Context().Value(ctxKey).(string)
	if referenceID == "" {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - Reset_Password - Execution completed in ", duration)
	}()

	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	param, _ := utils.Request(r)
	logger.Info(referenceID, "INFO - ResetPassword - params: ", param)

	email, ok := param["email"].(string)
	if !ok || email == "" {
		logger.Error(referenceID, "ERROR - Reset_Password - Missing email")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	redisClient := rds.GetRedisClient()
	if redisClient == nil {
		logger.Error(referenceID, "ERROR - ResetPassword - Redis client is not initialized")
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	if ttl, err := utils.SendMailLimiter(redisClient, referenceID, email, "Reset Password", time.Duration(configs.GetOTPExpireTime())*time.Second); err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password - ", err)
		result.ErrorCode = "429001"
		result.ErrorMessage = fmt.Sprintf("%s. Please try again in %d seconds", err.Error(), int(ttl.Seconds()))
		result.Payload["remaining_time"] = int(ttl.Seconds())
		utils.Response(w, result)
		return
	}

	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password - Failed to get DB connection: ", err)
		result.ErrorCode = "500001"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	var emailFromDb string
	err = conn.QueryRow(`SELECT email FROM sysuser.user WHERE email = $1`, email).Scan(&emailFromDb)
	if err == sql.ErrNoRows {
		result.ErrorCode = "401001"
		result.ErrorMessage = "Unauthorized"
		utils.Response(w, result)
		return
	} else if err != nil {
		logger.Error(referenceID, "ERROR - ResetPassword - Query failed: ", err)
		result.ErrorCode = "500002"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	nonce, err := utils.RandomStringGenerator(8)
	if err != nil {
		logger.Error(referenceID, "ERROR - ResetPassword - Failed to generate nonce: ", err)
		result.ErrorCode = "500005"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	URLExpireTstamp := time.Now().Unix() + int64(configs.GetResetPassExpTime())
	message := fmt.Sprintf("%d|%s", URLExpireTstamp, email)
	urlSignature, err := crypto.GenerateHMAC(message, nonce)
	if err != nil {
		logger.Error(referenceID, "ERROR - ResetPassword - Failed to generate URL signature: ", err)
		result.ErrorCode = "500006"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	// Encode & Inject expiry ke dalam signature
	finalSignature := encodeAndInjectExpiry(urlSignature, URLExpireTstamp)

	logger.Info(referenceID, "INFO - ResetPassword - URL expire timestamp: ", URLExpireTstamp)
	logger.Info(referenceID, "INFO - ResetPassword - nonce (key): ", nonce)
	logger.Info(referenceID, "INFO - ResetPassword - message: ", message)
	logger.Info(referenceID, "INFO - ResetPassword - URL signature: ", urlSignature)
	logger.Info(referenceID, "INFO - ResetPassword - Final Signature: ", finalSignature)

	expiry := time.Duration(configs.GetResetPassExpTime()) * time.Second
	if err := redisClient.Set(context.Background(), fmt.Sprintf("url_signature:%s", finalSignature), message, expiry).Err(); err != nil {
		logger.Error(referenceID, "ERROR - ResetPassword - Failed to store URL in Redis: ", err)
		result.ErrorCode = "500007"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	logger.Info(referenceID, "INFO - ResetPassword - expire time: ", expiry)

	// Susun URL dengan signature + nonce
	clientURL := fmt.Sprintf("%s/reset-password-confirm/%s", configs.GetClientURL(), finalSignature+nonce)

	logger.Info(referenceID, "INFO - ResetPassword - clientURL: ", clientURL)

	err = mail.SendEmail(email, "Reset Password Request", fmt.Sprintf("Your Reset Password URL: %s\nThis will expire in %.0f minutes.", clientURL, expiry.Minutes()))
	if err != nil {
		logger.Error(referenceID, "ERROR - ResetPassword - Failed to send email: ", err)
		result.ErrorCode = "500008"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	result.Payload["status"] = "success"
	utils.Response(w, result)
}

func Reset_Password_Verify_URL(w http.ResponseWriter, r *http.Request) {
	var ctxKey HTTPContextKey = "requestID"
	referenceID, _ := r.Context().Value(ctxKey).(string)
	if referenceID == "" {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - Reset_Password_Verify_URL - Execution completed in ", duration)
	}()

	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	param, _ := utils.Request(r)

	logger.Info(referenceID, "INFO - Reset_Password_Verify_URL - params: ", param)

	newPassword, ok := param["new_password"].(string)
	if !ok || newPassword == "" || len(newPassword) < 8 || len(newPassword) > 30 {
		logger.Error(referenceID, "ERROR - Reset_Password_Verify_URL - Missing or invalid new password")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	//b71619a3b2ff744719f22fcd20dae490f9e4f898bd7ae72faa60377fb9827ac2
	//b71619a3b2ff744719f22fcd20dae490f9e4f898bd7ae72faa60377fb9827ac2

	urlSignature, ok := param["url_signature"].(string)
	if !ok || urlSignature == "" {
		logger.Error(referenceID, "ERROR - Reset_Password_Verify_URL - Missing url_signature")
		result.ErrorCode = "400002"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	logger.Info(referenceID, "INFO - Reset_Password_Verify_URL - url_signature from client: ", urlSignature)

	redisClient := rds.GetRedisClient()
	if redisClient == nil {
		logger.Error(referenceID, "ERROR - Reset_Password_Verify_URL - Redis client is not initialized")
		result.ErrorCode = "500001"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	signatureKey := fmt.Sprintf("url_signature:%s", urlSignature)
	storedMessage, err := redisClient.Get(context.Background(), signatureKey).Result()
	if err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password_Verify_URL - Invalid or expired signature")
		result.ErrorCode = "401002"
		result.ErrorMessage = "Unauthorized"
		utils.Response(w, result)
		return
	}

	var expireTstamp int64
	var email string
	fmt.Sscanf(storedMessage, "%d|%s", &expireTstamp, &email)
	if time.Now().Unix() > expireTstamp {
		logger.Error(referenceID, "ERROR - Reset_Password_Verify_URL - Reset link expired")
		result.ErrorCode = "410001"
		result.ErrorMessage = "Gone"
		utils.Response(w, result)
		return
	}

	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password_Verify_URL - Failed to get DB connection: ", err)
		result.ErrorCode = "500002"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	salt, _ := utils.RandomStringGenerator(16)
	hashedPassword, _ := crypto.GeneratePBKDF2(newPassword, salt, 32, configs.GetPBKDF2Iterations())
	_, err = conn.Exec(`UPDATE sysuser."user" SET saltedpassword = $1, salt = $2 WHERE email = $3`, hashedPassword, salt, email)
	if err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password_Verify_URL - Failed to update password: ", err)
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	redisClient.Del(context.Background(), signatureKey)
	result.Payload["status"] = "success"
	utils.Response(w, result)
}
