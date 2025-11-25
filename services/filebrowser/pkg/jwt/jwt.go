package jwt

import (
	"errors"

	"github.com/golang-jwt/jwt/v4"
)

var (
	secretKey []byte
)

// InitializeJWT initializes JWT with secret key
func InitializeJWT(secret string) {
	secretKey = []byte(secret)
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// VerifyToken verifies and parses a JWT token
func VerifyToken(tokenString string) (*Claims, error) {
	if len(secretKey) == 0 {
		return nil, errors.New("JWT not initialized")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

