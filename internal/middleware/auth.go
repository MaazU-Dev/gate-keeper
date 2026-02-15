package middleware

import (
	"context"
	"gate-keeper/internal/auth"
	"gate-keeper/internal/config"
	"net/http"
)

func AuthMiddleware(next http.Handler, authTokenSecret string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		userID, err := auth.ValidateJWT(authToken, authTokenSecret)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), config.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
