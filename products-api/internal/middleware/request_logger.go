package middleware

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder keeps track of the status code written by the handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs method, path, status and latency for each HTTP request.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		started := time.Now()
		next.ServeHTTP(recorder, r)
		duration := time.Since(started)
		log.Printf("%s %s -> %d (%s)", r.Method, r.URL.Path, recorder.status, duration)
	})
}
