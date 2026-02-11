package main

import (
	"gate-keeper/internal/auth"
	"net/http"
)

func (cfg *Config) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		userID, err := auth.ValidateJWT(authToken, cfg.AuthTokenSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		r.Header.Set("X-User-ID", userID)
		next.ServeHTTP(w, r)
	})
}
