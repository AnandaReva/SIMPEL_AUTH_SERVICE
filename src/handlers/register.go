package handlers

import (
	"auth_service/configs"
	"auth_service/crypto"
	"auth_service/db"
	"auth_service/logger"
	"auth_service/mail"
	"auth_service/rds"
	"strings"

	"auth_service/utils"
	"context"
	"fmt"
	"net/http"
	"time"
)

/*
{
	"username" : "master",
	"full_name" : "Master User",
	"password" : "master123"
}
*/

func Register(w http.ResponseWriter, r *http.Request) {
	var ctxKey HTTPContextKey = "requestID"
	referenceID, _ := r.Context().Value(ctxKey).(string)
	if referenceID == "" {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - Register - Execution completed in ", duration)
	}()

	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	param, _ := utils.Request(r)

	logger.Info(referenceID, "INFO - Register - params: ", param)

	// Validasi input
	username, ok := param["username"].(string)
	if !ok || username == "" || len(username) < 6 {
		logger.Error(referenceID, "ERROR - Register - Missing username")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	email, ok := param["email"].(string)
	if !ok || email == "" {
		logger.Error(referenceID, "ERROR - Register - Missing email")
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
	if !ok || password == "" || len(password) < 8 {
		logger.Error(referenceID, "ERROR - Register - Missing password")
		result.ErrorCode = "400003"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to get DB connection: ", err)
		result.ErrorCode = "500002"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	// cek apakah email atau username sudah ada
	var existingField string
	queryCheck := `SELECT CASE WHEN EXISTS (SELECT 1 FROM sysuser."user" WHERE username = $1) THEN 'username' WHEN EXISTS (SELECT 1 FROM sysuser."user" WHERE email = $2) THEN 'email' ELSE NULL END AS existing_field;`
	errCheck := conn.Get(&existingField, queryCheck, username, email)
	if errCheck == nil && existingField != "" {
		result.ErrorCode = "409001"
		result.ErrorMessage = fmt.Sprintf("%s already exists", existingField)
		utils.Response(w, result)
		return
	}

	redisClient := rds.GetRedisClient()
	if redisClient == nil {
		logger.Error(referenceID, "ERROR - Register - Redis client is not initialized")
		result.ErrorCode = "500002"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	ttl, err := utils.SendMailLimiter(redisClient, referenceID, email, "Registration OTP", time.Duration(configs.GetOTPExpireTime())*time.Second)
	if err != nil {
		logger.Error(referenceID, "ERROR - Register - ", err)
		result.ErrorCode = "429001"
		result.ErrorMessage = fmt.Sprintf("%s. Please try again in %d seconds", err.Error(), int(ttl.Seconds()))
		utils.Response(w, result)
		return
	}

	// Cek apakah email ini sudah memiliki OTP yang valid
	redisKey := fmt.Sprintf("otp_active:%s", email)
	exists, err := redisClient.Exists(context.Background(), redisKey).Result()
	if err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to check Redis: ", err)
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	if exists > 0 {
		result.ErrorCode = "429001"
		result.ErrorMessage = "OTP already sent. Please wait before requesting again."
		utils.Response(w, result)
		return
	}

	otpInt, err := utils.RandoNnumberGenerator(6)
	otp := fmt.Sprintf("%06d", otpInt) // Pastikan selalu 6 digit dengan padding nol jika perlu

	if err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to generate OTP: ", err)
		result.ErrorCode = "500001"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}
	logger.Info(referenceID, "INFO - Register - OTP generated: ", otp)

	message := fmt.Sprintf("%s|%s|%s|%s", email, fullName, password, username)
	logger.Info(referenceID, "INFO - Register - message: ", message)
	logger.Info(referenceID, "INFO - Register - key (otp): ", otp)

	otpSignature, err := crypto.GenerateHMAC(message, otp)
	if err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to generate OTP signature: ", err)
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return

	}

	logger.Info(referenceID, "INFO - Register - OTP signature: ", otpSignature)

	redisOTPKey := fmt.Sprintf("otp_signature:%s", otpSignature)

	expiry := time.Duration(configs.GetOTPExpireTime()) * time.Second
	if err := redisClient.Set(context.Background(), redisOTPKey, message, expiry).Err(); err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to store OTP in Redis: ", err)
		result.ErrorCode = "500004"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	// Send OTP via SMTP
	err = mail.SendEmail(email, "OTP Verification", fmt.Sprintf("Your OTP is: %s\nThis will expire in %.0f seconds.", otp, expiry.Seconds()))

	if err != nil {
		logger.Error(referenceID, "ERROR - Register - Failed to send OTP via email: ", err)
		result.ErrorCode = "500005"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	// count OTPExpireTstamp
	OTPExpireTstamp := time.Now().Unix() + int64(configs.GetOTPExpireTime())
	logger.Info(referenceID, "INFO - Register - calculated OTPExpireTstamp: ", OTPExpireTstamp)

	result.Payload["otp_expire_tstamp"] = OTPExpireTstamp
	result.Payload["status"] = "success"

	utils.Response(w, result)
}

// message = username+email+password, full_name
//  otp_signature = hmac-sha256( message , otp)
// key:value =>    username:{otp_signature+ message}

func Register_Verify_OTP(w http.ResponseWriter, r *http.Request) {
	var ctxKey HTTPContextKey = "requestID"
	referenceID, _ := r.Context().Value(ctxKey).(string)
	if referenceID == "" {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - Reg_Verify_OTP - Execution completed in ", duration)
	}()

	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	param, _ := utils.Request(r)

	logger.Info(referenceID, "INFO - Reg_Verify_OTP - params: ", param)

	otpSignature, _ := param["otp_signature"].(string)

	if otpSignature == "" {
		logger.Error(referenceID, "ERROR - Reg_Verify_OTP - Missing otp_signature")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	redisClient := rds.GetRedisClient()
	redisKey := fmt.Sprintf("otp_signature:%s", otpSignature)
	message, err := redisClient.Get(context.Background(), redisKey).Result()

	logger.Info(referenceID, "INFO - Reg_Verify_OTP - otp_signature: ", otpSignature)
	logger.Info(referenceID, "INFO - Reg_Verify_OTP - message from redis: ", message)

	if err != nil {
		logger.Error(referenceID, "ERROR - Reg_Verify_OTP - OTP not found in Redis: ", err)
		result.ErrorCode = "401002"
		result.ErrorMessage = "Unauthorized"
		utils.Response(w, result)
		return
	}

	// Memisahkan berdasarkan '|'
	parts := strings.Split(message, "|")

	// Pastikan jumlah bagian sesuai sebelum mengaksesnya
	if len(parts) != 4 {
		logger.Error(referenceID, "ERROR - Reg_Verify_OTP - Invalid data format in Redis")
		result.ErrorCode = "500007"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	email := parts[0]
	fullName := parts[1]
	password := parts[2]
	username := parts[3]

	logger.Info(referenceID, "INFO - Reg_Verify_OTP - username: ", username)
	logger.Info(referenceID, "INFO - Reg_Verify_OTP - email: ", email)
	logger.Info(referenceID, "INFO - Reg_Verify_OTP - password: ", password)
	logger.Info(referenceID, "INFO - Reg_Verify_OTP - full_name: ", fullName)

	salt, _ := utils.RandomStringGenerator(16)
	saltedPassword, _ := crypto.GeneratePBKDF2(password, salt, 32, configs.GetPBKDF2Iterations())

	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - Reg_Verify_OTP - Failed to get DB connection: ", err)
		result.ErrorCode = "500002"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	queryToRegister := `INSERT INTO sysuser.user (username, full_name, email, st, salt, saltedpassword, data, role) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id;`
	var newUserId int
	err = conn.Get(&newUserId, queryToRegister, username, fullName, email, 1, salt, saltedPassword, "{}", "system user")
	if err != nil {
		logger.Error(referenceID, "ERROR - Reg_Verify_OTP - Failed to insert new account: ", err)
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}
	logger.Info(referenceID, "INFO - Reg_Verify_OTP - New user ID: ", newUserId)

	redisClient.Del(context.Background(), redisKey)

	result.Payload["success"] = "success"
	utils.Response(w, result)
}
