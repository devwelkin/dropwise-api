package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs details about HTTP requests including method, path,
// status code, and request duration
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Start timer
		start := time.Now()

		// Create a custom response writer to capture the status code
		crw := &customResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Call the next handler
		next(crw, r)

		// Calculate duration
		duration := time.Since(start)

		// Log request details
		log.Printf(
			"[%s] %s %s - Status: %d - Duration: %v",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
			crw.statusCode,
			duration,
		)
	}
}

// customResponseWriter is a wrapper around http.ResponseWriter that captures the status code
type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code before calling the underlying ResponseWriter
func (crw *customResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}
