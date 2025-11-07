package jwt

import (
	"errors"
	"time"

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

// GenerateToken generates a new JWT token
func GenerateToken(userID, email, role string, expirationTime time.Duration) (string, error) {
	if len(secretKey) == 0 {
		return "", errors.New("JWT not initialized")
	}

	expirationTimeFromNow := time.Now().Add(expirationTime)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTimeFromNow),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "ar-13-backend",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
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

// GenerateRefreshToken generates a refresh token (longer expiration)
func GenerateRefreshToken(userID string) (string, error) {
	return GenerateToken(userID, "", "", 30*24*time.Hour) // 30 days
}

