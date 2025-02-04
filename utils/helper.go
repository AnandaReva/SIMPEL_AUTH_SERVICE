package utils

import (
	"auth_service/logger"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func JSONencode(data any) (string, error) {
	var buffer bytes.Buffer
	// Buat encoder yang menulis ke buffer
	encoder := json.NewEncoder(&buffer)
	// Set agar encoder tidak melakukan escaping HTML
	encoder.SetEscapeHTML(false)
	// Encode data ke JSON dan simpan ke buffer
	if err := encoder.Encode(data); err != nil {
		fmt.Println("Error encoding JSON:", err)
		return "", err
	}
	// Mendapatkan hasil JSON sebagai string
	jsonString := buffer.String()
	return jsonString, nil
}
func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

// format response
type ResultFormat struct {
	ErrorCode    string
	ErrorMessage string
	Payload      map[string]any
}

func Response(w http.ResponseWriter, result ResultFormat) {
	// Get the first 3 digits from ErrorCode (e.g., "500003" -> "500")
	var httpErrCode int

	if len(result.ErrorCode) >= 3 {
		// Extract the first 3 digits of the ErrorCode
		_, err := fmt.Sscanf(result.ErrorCode[:3], "%d", &httpErrCode)
		if err != nil {
			httpErrCode = http.StatusInternalServerError
		}
	} else {
		httpErrCode = http.StatusInternalServerError
	}

	// Handle special cases for 000 (OK status)
	if result.ErrorCode[:3] == "000" {
		httpErrCode = http.StatusOK
	}

	// Set HTTP status code based on the extracted error code (401, 400, 500, etc.)
	if httpErrCode == 0 {
		httpErrCode = http.StatusInternalServerError
	}

	// Set the response content type and status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErrCode)

	// Encode the result as JSON and send it in the response body
	if err := json.NewEncoder(w).Encode(result); err != nil {
		logger.Error("Unknown", "ERROR - Response encoding failed: ", err)
	}
}

func Request(r *http.Request) (map[string]any, error) {
	var data map[string]any
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	// Build the log string for parameters
	var logParams []string
	for key, value := range data {
		// Format the value as a string (use quotes for string values)
		var formattedValue string
		switch v := value.(type) {
		case string:
			formattedValue = fmt.Sprintf("\"%s\"", v) // Quote string values
		default:
			formattedValue = fmt.Sprintf("%v", v) // For other types, just use the default format
		}

		// Append to the log array
		logParams = append(logParams, fmt.Sprintf("%s : %s", key, formattedValue))
	}

	// Join all parameters into a single string
	logMessage := fmt.Sprintf("INFO - Received parameters: [%s]", strings.Join(logParams, ", "))
	logger.Info("Request", logMessage)

	return data, nil
}
