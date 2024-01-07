package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
)

var invalidToken error = errors.New("invalid token")

func CheckCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("simple_stack_token")
		if errors.Is(err, http.ErrNoCookie) {
			http.Error(w, fmt.Sprintf("failed to authenticate: %v", err), http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, fmt.Sprintf("failed to get cookie: %v", err), http.StatusInternalServerError)
			return
		}

		if cookie.Value != os.Getenv("SIMPLE_STACK_TOKEN") {
			http.Error(w, fmt.Sprintf("failed to authenticate: %v", invalidToken), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
