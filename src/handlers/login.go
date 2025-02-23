package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"auth_service/crypto"
	"auth_service/db"
	"auth_service/logger"
	"auth_service/utils"
)

/*
 \d sysuser.user;
                                           Table "sysuser.user"
     Column     |          Type          | Collation | Nullable |                 Default
----------------+------------------------+-----------+----------+------------------------------------------
 username       | character varying(30)  |           | not null |
 full_name      | character varying(128) |           | not null |
 st             | integer                |           | not null |
 salt           | character varying(64)  |           | not null |
 saltedpassword | character varying(128) |           | not null |
 data           | jsonb                  |           | not null |
 id             | bigint                 |           | not null | nextval('sysuser.user_id_seq'::regclass)
 role           | character varying(128) |           | not null |
Indexes:
    "user_pkey" PRIMARY KEY, btree (id)
    "user_unique_name" UNIQUE CONSTRAINT, btree (username)
Referenced by:
    TABLE "sysuser.token" CONSTRAINT "fk_user_id" FOREIGN KEY (user_id) REFERENCES sysuser."userCred"(id) ON DELETE CASCADE



tubes=> \d sysuser.session;
                         Table "sysuser.session"
     Column     |          Type          | Collation | Nullable | Default
----------------+------------------------+-----------+----------+---------
 session_id     | character varying(16)  |           | not null |
 user_id        | bigint                 |           | not null |
 session_hash   | character varying(128) |           | not null |
 tstamp         | bigint                 |           | not null |
 st             | integer                |           | not null |
 last_ms_tstamp | bigint                 |           |          |
 last_sequence  | bigint                 |           |          |
Indexes:
    "session_pkey" PRIMARY KEY, btree (session_id)
    "session_user_id_key" UNIQUE CONSTRAINT, btree (user_id)

\d sysuser.token;
                      Table "sysuser.token"
 Column  |         Type          | Collation | Nullable | Default
---------+-----------------------+-----------+----------+---------
 user_id | bigint                |           | not null |
 token   | character varying(16) |           | not null |
 tstamp  | bigint                |           | not null |
Indexes:
    "token_pkey" PRIMARY KEY, btree (user_id, token)
    "unique_user_id" UNIQUE CONSTRAINT, btree (user_id)
Foreign-key constraints:
    "fk_user_id" FOREIGN KEY (user_id) REFERENCES sysuser."user"(id) ON DELETE CASCADE
*/

// GenerateNonce membuat nonce acak sepanjang 8 byte
func GenerateNonce() (string, error) {
	nonce, err := utils.RandomStringGenerator(8)
	if err != nil {
		return "", errors.New("failed to generate nonce")
	}
	return nonce, nil
}

type UserCred struct {
	ID             int64  `db:"id"`
	Salt           string `db:"salt"`
	SaltedPassword string `db:"saltedpassword"`
}

/* type UserData struct {
	Username    string                 `db:"username"`
	FullName    string                 `db:"full_name"`
	Data        map[string]interface{} `db:"data"` // jsonb
	SessionID   string                 `db:"session_id"`
	SessionHash string                 `db:"session_hash"`
}
*/

// !!!NOTE : DONT GIVE ANY DETAILED ERROR MESSAGE TO CLIENT

/*
### Langkah-langkah Login dengan Verifikasi Token
*/
func Login(w http.ResponseWriter, r *http.Request) {
	var ctxKey HTTPContextKey = "requestID"
	referenceID, ok := r.Context().Value(ctxKey).(string)
	if !ok {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - Login - Execution completed in ", duration)
	}()

	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	param, _ := utils.Request(r)

	username, ok := param["username"].(string)
	if !ok || username == "" {
		logger.Error(referenceID, "ERROR - Login - Missing username")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	password, ok := param["password"].(string)
	if !ok || password == "" {
		logger.Error(referenceID, "ERROR - Login - Missing password")
		result.ErrorCode = "400003"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	halfNonce, ok := param["half_nonce"].(string)
	if !ok || len(halfNonce) < 8 {
		logger.Error(referenceID, "ERROR - Login - Missing half_nonce")
		result.ErrorCode = "400004"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - Login - DB connection failed: ", err)
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	var userCred UserCred
	queryGetUser := `SELECT id, salt, saltedpassword FROM sysuser.user WHERE username = $1`
	if err := conn.Get(&userCred, queryGetUser, username); err != nil {
		logger.Error(referenceID, "ERROR - Login - User not found: ", err)
		result.ErrorCode = "401000"
		result.ErrorMessage = "Unauthorized"
		utils.Response(w, result)
		return
	}

	halfNonce2, errHNC2 := GenerateNonce()

	fullNonce := halfNonce + halfNonce2
	token, errTkn := crypto.GenerateHMAC(userCred.SaltedPassword, fullNonce)

	if errHNC2 != nil || errTkn != nil {
		logger.Error(referenceID, "ERROR - Login - GenerateNonce or GenerateHMAC token failed generation failed", errTkn, errHNC2)
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	logger.Info(referenceID, "INFO - Login - fullNonce:", fullNonce)
	logger.Info(referenceID, "INFO - Login - token:", token)

	queryUpsertToken := `
		INSERT INTO sysuser.token (user_id, token, tstamp) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id) 
		DO UPDATE SET token = EXCLUDED.token, tstamp = EXCLUDED.tstamp`
	if _, err := conn.Exec(queryUpsertToken, userCred.ID, token, time.Now().Unix()); err != nil {
		logger.Error(referenceID, "ERROR - Login - Token upsert failed", err)
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	result.Payload["full_nonce"] = fullNonce
	result.Payload["salt"] = userCred.Salt
	utils.Response(w, result)
}

/*
	CLIENT SIDE AFTER LOGIN
	1. get full_nonce and salt
	2. client will need to craft token with full_nonce, salt and password -> token = hmac-sha256(SaltedPassword , fullNonce) , saltedPassword = argon2(password, salt)
	3. send token verify-token to be verified

*/

type UserData struct {
	Username string          `db:"username"`
	FullName string          `db:"full_name"`
	Role     string          `db:"role"`
	Data     json.RawMessage `db:"data"` // jsonb
}

func Verify_Token(w http.ResponseWriter, r *http.Request) {
	var ctxKey HTTPContextKey = "requestID"
	referenceID, ok := r.Context().Value(ctxKey).(string)
	if !ok {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - VerifyToken - Execution completed in", duration)
	}()

	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	param, _ := utils.Request(r)
	tokenClient, ok := param["token"].(string)
	if !ok || tokenClient == "" {
		logger.Error(referenceID, "ERROR - VerifyToken - Missing token")
		result.ErrorCode = "400001"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - VerifyToken - DB connection failed")
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	var userID int64
	var storedToken string
	var tokenCreatedStamp int64
	queryGetToken := `SELECT user_id, token, tstamp FROM sysuser.token WHERE token = $1`
	err = conn.QueryRow(queryGetToken, tokenClient).Scan(&userID, &storedToken, &tokenCreatedStamp)
	if err != nil {
		logger.Error(referenceID, "ERROR - VerifyToken - Invalid token", err)
		result.ErrorCode = "401000"
		result.ErrorMessage = "Unauthorized"
		utils.Response(w, result)
		return
	}

	//check time for validation
	timeForValidation := time.Now().Unix() - tokenCreatedStamp
	logger.Info(referenceID, "ERROR - VerifyToken - Time for validation (s): ", timeForValidation)
	if timeForValidation > 100 { // Token expiry period (e.g., 100 seconds)
		logger.Error(referenceID, "ERROR - VerifyToken - Token Expired (> )")
		result.ErrorCode = "401000"
		result.ErrorMessage = "Unauthorized"
		utils.Response(w, result)
		return
	}

	// Delete token after validation
	queryDeleteToken := `DELETE FROM sysuser.token WHERE user_id = $1 AND token = $2`
	if _, err := conn.Exec(queryDeleteToken, userID, tokenClient); err != nil {
		logger.Warning(referenceID, "WARNING - VerifyToken - Token cleanup failed", err)
	}

	// Generate session ID and session hash
	sessionID, errMsg := utils.RandomStringGenerator(16)
	if errMsg != nil {
		logger.Error(referenceID, "ERROR - VerifyToken - Session ID generation failed")
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	sessionHash, errMsg := crypto.GenerateHMAC(tokenClient, sessionID)
	if errMsg != nil {
		logger.Error(referenceID, "ERROR - VerifyToken - HMAC computation failed")
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	// Create or update session
	currentTime := time.Now().Unix()
	queryUpsertSession := `
		INSERT INTO sysuser.session (session_id, user_id, session_hash, tstamp, st)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			session_id = EXCLUDED.session_id,
			session_hash = EXCLUDED.session_hash,
			tstamp = EXCLUDED.tstamp,
			st = EXCLUDED.st`

	if _, err := conn.Exec(queryUpsertSession, sessionID, userID, sessionHash, currentTime, 1); err != nil {
		logger.Error(referenceID, "ERROR - VerifyToken - Session upsert failed")
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal server error"
		utils.Response(w, result)
		return
	}

	// Fetch user data
	var userData UserData
	queryGetUserData := `SELECT username, full_name, role, COALESCE(data, '{}'::jsonb) AS data FROM sysuser.user WHERE id = $1`
	if err := conn.Get(&userData, queryGetUserData, userID); err != nil {
		logger.Error(referenceID, "ERROR - VerifyToken - User not found", err)
		result.ErrorCode = "401000"
		result.ErrorMessage = "Unauthorized"
		utils.Response(w, result)
		return
	}

	// Prepare response payload
	result.Payload["session_id"] = sessionID
	result.Payload["session_hash"] = sessionHash
	result.Payload["username"] = userData.Username
	result.Payload["full_name"] = userData.FullName
	result.Payload["role"] = userData.Role
	result.Payload["data"] = userData.Data

	utils.Response(w, result)
}
