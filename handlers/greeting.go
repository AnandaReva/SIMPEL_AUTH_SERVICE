package handlers

import (
	"auth_service/logger"
	"auth_service/utils"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPContextKey string

// RequestBody adalah struktur untuk membaca JSON dari body request.
type RequestBody struct {
	Name string `json:"name"`
}

// Greeting handles requests to the root endpoint ("/").
func Greeting(w http.ResponseWriter, r *http.Request) {
	var res string

	// Context key untuk request ID
	var ctxKey HTTPContextKey = "requestID"
	referenceID, ok := r.Context().Value(ctxKey).(string)
	if !ok {
		referenceID = "unknown"
	}

	// Set header untuk response JSON
	w.Header().Set("Content-Type", "application/json")

	// Default response
	response := map[string]any{
		"error_code":    "000000000",
		"error_message": "",
	}

	// Validasi metode HTTP
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		logger.Error(referenceID, "Invalid method: ", r.Method)
		response["error_code"] = "405000001"
		response["error_message"] = "Method Not Allowed"
		res, _ = utils.JSONencode(response)
		http.Error(w, res, http.StatusMethodNotAllowed)
		return
	}

	// Penanganan untuk metode GET
	if r.Method == http.MethodGet {
		res := "Hello!"
		fmt.Fprint(w, res)
		return
	}

	// Penanganan untuk metode POST
	logger.Info(referenceID, "Request Method: POST")

	// Periksa apakah Content-Type adalah application/json
	if contentType := r.Header.Get("Content-Type"); contentType != "application/json" {
		logger.Error(referenceID, "Invalid content-type: ", contentType)
		response["error_code"] = "400000001"
		response["error_message"] = "Bad Request. Not JSON"
		res, _ = utils.JSONencode(response)
		http.Error(w, res, http.StatusBadRequest)
		return
	}

	// Baca body POST
	body, err := io.ReadAll(r.Body)
	if err != nil {
		logger.Error(referenceID, "Failed to read body: ", err)
		response["error_code"] = "400000002"
		response["error_message"] = "Bad Request. Can't read POST body"
		res, _ = utils.JSONencode(response)
		http.Error(w, res, http.StatusBadRequest)
		return
	}
	logger.Info(referenceID, "POST Body: ", string(body))

	// Decode JSON body ke struct RequestBody
	var req RequestBody
	err = json.Unmarshal(body, &req)
	if err != nil {
		logger.Error(referenceID, "Failed to decode JSON: ", err)
		response["error_code"] = "400000003"
		response["error_message"] = "Bad Request. Invalid JSON"
		res, _ = utils.JSONencode(response)
		http.Error(w, res, http.StatusBadRequest)
		return
	}

	// Buat message berdasarkan input
	if req.Name == "" {
		response["message"] = "Hello!"
	} else {
		response["message"] = "Hello, " + req.Name + "!"
	}
	response["reference_id"] = referenceID

	// Encode response ke JSON dan kirimkan
	res, _ = utils.JSONencode(response)
	logger.Info(referenceID, "Response body: ", res)
	fmt.Fprint(w, res)
}
