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
	"fmt"
	"net/http"
	"time"
)



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

	email, _ := param["email"].(string)
	if email == "" {
		logger.Error(referenceID, "ERROR - Reset_Password - Missing email")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}


	nonce, _ := param["nonce"].(string)

	if nonce == "" {
		logger.Error(referenceID, "ERROR - Reset_Password - Missing nonce")
		result.ErrorCode = "400002"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}


	// Cek apakah email sudah terdaftar

	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password - Failed to get DB connection: ", err)
		result.ErrorCode = "500001"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	// cek apakah email terdaftar ada
	var emailFromDb string
	queryCheck := `SELECT email FROM sysuser."user" WHERE email = $1`
	errCheck := conn.Get(&emailFromDb, queryCheck,  email)
	if errCheck != nil  {
		result.ErrorCode = "409001"
		result.ErrorMessage = ""
		utils.Response(w, result)
		return
	}

	//jika ada kirim OTP ke email
	redisClient := rds.GetRedisClient()
	if redisClient == nil {
		logger.Error(referenceID, "ERROR - Reset_Password - Redis client is not initialized")
		result.ErrorCode = "500002"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	if ttl, err := utils.OTPRateLimiter(redisClient, email); err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password - ", err)
		result.ErrorCode = "429002"
		result.ErrorMessage = fmt.Sprintf("%s. Please try again in %d seconds", err.Error(), int(ttl.Seconds()))
		utils.Response(w, result)
		return
	}

	// Cek apakah email ini sudah memiliki OTP yang valid
	redisKey := fmt.Sprintf("otp_active:%s", email)
	exists, err := redisClient.Exists(context.Background(), redisKey).Result()
	if err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password - Failed to check Redis: ", err)
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
		logger.Error(referenceID, "ERROR - Reset_Password - Failed to generate OTP: ", err)
		result.ErrorCode = "500001"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}
	logger.Info(referenceID, "INFO - Reset_Password - OTP generated: ", otp)

	message := fmt.Sprintf("%s|%s", nonce, email)
	logger.Info(referenceID, "INFO - Reset_Password - message: ", message)
	logger.Info(referenceID, "INFO - Reset_Password - key (otp): ", otp)

	otpSignature, err := crypto.GenerateHMAC(message, otp)
	if err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password - Failed to generate OTP signature: ", err)
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return

	}

	logger.Info(referenceID, "INFO - Reset_Password - OTP signature: ", otpSignature)

	redisOTPKey := fmt.Sprintf("otp_signature:%s", otpSignature)

	expiry := time.Duration(configs.GetOTPExpireTime()) * time.Second
	if err := redisClient.Set(context.Background(), redisOTPKey, message, expiry).Err(); err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password - Failed to store OTP in Redis: ", err)
		result.ErrorCode = "500004"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	// Send OTP via SMTP
	err = mail.SendEmail(email, "OTP Verification for Reset Password", fmt.Sprintf("Your OTP is: %s\nThis will expire in %.0f seconds.", otp, expiry.Seconds()))

	if err != nil {
		logger.Error(referenceID, "ERROR - Reset_Password - Failed to send OTP via email: ", err)
		result.ErrorCode = "500005"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	// count OTPExpireTstamp
	OTPExpireTstamp := time.Now().Unix() + int64(configs.GetOTPExpireTime())
	logger.Info(referenceID, "INFO - Reset_Password - calculated OTPExpireTstamp: ", OTPExpireTstamp)

	result.Payload["otp_expire_tstamp"] = OTPExpireTstamp
	result.Payload["status"] = "success"

	utils.Response(w, result)
	



}



func Reset_Password_Verify_OTP(w http.ResponseWriter, r *http.Request) {

	var ctxKey HTTPContextKey = "requestID"
	referenceID, _ := r.Context().Value(ctxKey).(string)
	if referenceID == "" {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - Reset_Password_Verify_OTP - Execution completed in ", duration)
	}()

	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	param, _ := utils.Request(r)

	email, _ := param["otp_signature"].(string)
	if email == "" {
		logger.Error(referenceID, "ERROR - Reset_Password_Verify_OTP - Missing email")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}





}