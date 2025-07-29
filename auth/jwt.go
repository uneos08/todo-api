package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key") // Лучше читать из env

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// CreateJWTToken создаёт access JWT для пользователя с указанным ID
func CreateJWTToken(userID int) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Minute)), // Access token на 15 мин
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "todo-api-access_token",
			Subject:   "access_token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// CreateRefreshToken создаёт refresh JWT для пользователя с указанным ID
func CreateRefreshToken(userID int) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "refresh",
			Subject:   "refresh_token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseJWTToken парсит и проверяет JWT, возвращая claims
func ParseJWTToken(tokenStr string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}
