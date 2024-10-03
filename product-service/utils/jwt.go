package utils

import (
	"os"

	"github.com/dgrijalva/jwt-go"
)

var JwtKey = []byte(getEnv("JWT_SECRET_KEY", "default_secret_key"))

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}
