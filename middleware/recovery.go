package middleware

import (
	"fmt"
	"net/http"
)

func Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				fmt.Printf("Recovered from panic: %v\n", err)
			}
		}()
		next(w, r)
	}
}
