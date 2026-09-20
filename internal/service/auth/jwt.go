package auth

import (
	"chesscourse/internal/apperror"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"` // добавили
	jwt.RegisteredClaims
}

var jwtSecretKey = initJWTSecret()

func initJWTSecret() []byte {
	godotenv.Load()
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}
	return []byte(secret)
}

func GenerateToken(userID int, role string) (string, error) {
	expirationDate := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID: userID,
		Role:   role, // добавили
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationDate),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecretKey)
}

func ValidateToken(tokenStr string) (int, string, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		return jwtSecretKey, nil
	})

	if err != nil {
		return 0, "", apperror.ErrInvalidToken
	}

	if !token.Valid {
		return 0, "", apperror.ErrInvalidToken
	}

	return claims.UserID, claims.Role, nil
}
