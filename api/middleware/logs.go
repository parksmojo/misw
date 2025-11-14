package middleware

import (
	"log"
	"net/http"
	"time"
)

func LogRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf("Recieved %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next(lrw, r)

		duration := time.Since(start)
		log.Printf("Finished %s %s in %v with code %d", r.Method, r.URL.Path, duration, lrw.statusCode)
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
