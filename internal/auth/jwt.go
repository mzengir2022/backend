package auth

import (
	"my-project/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

func InitializeJWT(cfg *config.Config) {
	jwtKey = []byte(cfg.JWTSecret)
}

type Claims struct {
	UserID      uint   `json:"user_id"`
	PhoneNumber string `json:"phone_number"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateJWT(userID uint, phoneNumber, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:      userID,
		PhoneNumber: phoneNumber,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ValidateJWT(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}
