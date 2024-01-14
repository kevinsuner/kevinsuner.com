package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

var errInvalidToken error = errors.New("invalid token")

func CheckCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("admin_token")
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, fmt.Sprintf("failed to authenticate: %v", err), http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("failed to get cookie: %v", err), http.StatusInternalServerError)
			return
		}

		if cookie.Value != os.Getenv("ADMIN_TOKEN") {
			http.Error(w, fmt.Sprintf("failed to authenticate: %v", errInvalidToken), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
