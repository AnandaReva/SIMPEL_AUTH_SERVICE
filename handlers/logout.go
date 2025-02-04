package handlers

import (
	"auth_service/db"
	"auth_service/logger"
	"auth_service/utils"
	"net/http"
	"time"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	var ctxKey HTTPContextKey = "requestID"
	referenceID, ok := r.Context().Value(ctxKey).(string)
	if !ok {
		referenceID = "unknown"
	}

	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		logger.Debug(referenceID, "DEBUG - Logout - Execution completed in ", duration)
	}()

	result := utils.ResultFormat{
		ErrorCode:    "000000",
		ErrorMessage: "",
		Payload:      make(map[string]any),
	}

	param, _ := utils.Request(r)

	logger.Info(referenceID, "INFO - Logout - param:  ", param)
	// Validasi input
	sessionId, ok := param["session_id"].(string)
	if !ok || sessionId == "" {
		logger.Error(referenceID, "ERROR - Logout - Missing sessionId")
		result.ErrorCode = "400000"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	// Dapatkan koneksi database
	conn, err := db.GetConnection()
	if err != nil {
		logger.Error(referenceID, "ERROR - Logout - DB connection failed: ", err)
		result.ErrorCode = "500000"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	// Hapus sesi dari database
	queryToDeleteSession := `DELETE FROM sysuser.session WHERE session_id = $1`

	logger.Info(referenceID, "INFO - Logout - Executing query to delete session for session_id:", sessionId)
	res, err := conn.Exec(queryToDeleteSession, sessionId)
	if err != nil {
		logger.Error(referenceID, "ERROR - Logout - Failed to delete session: ", err)
		result.ErrorCode = "500001"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logger.Error(referenceID, "ERROR - Logout - Unable to get affected rows")
		result.ErrorCode = "500003"
		result.ErrorMessage = "Internal Server Error"
		utils.Response(w, result)
		return
	}

	if rowsAffected == 0 {
		logger.Warning(referenceID, "WARNING - Logout - No session found for session_id:", sessionId)
		result.ErrorCode = "400002"
		result.ErrorMessage = "Invalid request"
		utils.Response(w, result)
		return
	}

	// Berhasil logout
	result.Payload["status"] = "success"
	utils.Response(w, result)
}
