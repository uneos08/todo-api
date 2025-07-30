package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key")

var ErrNoAuthHeader = errors.New("authorization header is missing")
var ErrInvalidAuthHeader = errors.New("authorization header is invalid")

type CustomClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// Извлекает JWT из заголовка Authorization: Bearer <token>
func ExtractClaimsFromRequest(r *http.Request) (*CustomClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, ErrNoAuthHeader
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, ErrInvalidAuthHeader
	}

	tokenStr := parts[1]
	claims, err := ParseJWTToken(tokenStr)
	if err != nil {
		return nil, err
	}

	return claims, nil
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

// ExtractUserIDFromRequest извлекает userID из JWT в запросе
func ExtractUserIDFromRequest(r *http.Request) (int, error) {
	claims, err := ExtractClaimsFromRequest(r)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
