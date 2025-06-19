// internal/utils/jwt.go
package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
)

var jwtKey []byte
var jwtExpiration time.Duration

// Claims struct for JWT
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// LoadEnv loads environment variables from the .env file
func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("error loading .env file: %w", err)
	}

	// Set JWT secret key and expiration time from environment variables
	jwtKey = []byte(os.Getenv("JWT_SECRET"))

	expiration, err := time.ParseDuration(os.Getenv("JWT_EXPIRATION"))
	if err != nil {
		return fmt.Errorf("error parsing JWT_EXPIRATION: %w", err)
	}
	jwtExpiration = expiration

	return nil
}

// GenerateJWT generates a new JWT token
func GenerateJWT(userID, email string) (string, error) {
	expirationTime := time.Now().Add(jwtExpiration)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}

	return tokenString, nil
}

// VerifyJWT verifies a token and returns the claims
func VerifyJWT(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, err
	}

	return claims, nil
}
