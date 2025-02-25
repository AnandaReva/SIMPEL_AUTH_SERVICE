package middlewares

import (
	"auth_service/handlers"
	"auth_service/logger"
	"auth_service/utils"
	"context"
	"net/http"
	"strconv"
	"time"
)

func generateReferenceID(timer int64) string {
	timeBase36 := strconv.FormatUint(uint64(timer), 36)
	randString, err := utils.RandomStringGenerator(8)
	if err != nil {
		randString = "12345678"
	}
	reference_id := timeBase36 + "." + randString // concate
	return reference_id
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins
		w.Header().Set("Access-Control-Allow-Origin", "*")
		// Allow only GET and POST methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		// Allow only JSON content
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			// If preflight request, return 204 No Content
			w.WriteHeader(http.StatusNoContent)
			return
		}
		start := time.Now()
		requestID := generateReferenceID(start.UnixNano())

		// Log request details
		logger.Info(requestID, "Handle Request Started: ", r.Method, " ", r.URL.Path)
		logger.Info(requestID, "Query String: ", r.URL.RawQuery)
		logger.Info(requestID, "Headers:")
		for name, values := range r.Header {
			for _, value := range values {
				logger.Debug(requestID, name, ": ", value)
			}
		}

		// Add request ID to context ,
		// !!! note : context key is like mini state-management
		ctx := context.WithValue(r.Context(), handlers.HTTPContextKey("requestID"), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
		// Log completion and duration
		duration := time.Since(start)
		logger.Info(requestID, " Handle Request Completed in: ", duration)
		logger.Info(requestID, " ----------------------------------------------")

	})
}