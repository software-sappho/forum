package middleware

import (
	"fmt"
	"net/http"
	"time"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		srw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		start := time.Now()
		next.ServeHTTP(srw, r)

		fmt.Printf("[%s] %s %s - %d %s\n",
			start.Format("02-01-2006 15:04:05"),
			r.Method,
			r.URL.Path,
			srw.statusCode,
			time.Since(start),
		)
	})
}
