package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func ValidateJWT(tokenString, tokenSecret string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", errors.New("Invalid token")
	}
	tokenClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("Invalid token claims")
	}
	userID := fmt.Sprintf("%v", tokenClaims["user_id"])
	if userID == "" {
		return "", errors.New("User ID is required")
	}
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("You are not authenticated")
	}
	authToken := strings.Split(authHeader, " ")
	if len(authToken) < 2 || authToken[0] != "Bearer" {
		return "", errors.New("Invalid Bearer token format")
	}
	return authToken[1], nil
}
