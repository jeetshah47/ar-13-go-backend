package jwt

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

var (
	secretKey []byte
)

// InitializeJWT initializes JWT with secret key
func InitializeJWT(secret string) {
	secret = strings.TrimSpace(secret)
	secretKey = []byte(secret)
	// Debug: log secret key length and first few chars (for debugging only)
	if len(secret) > 0 {
		log.Printf("[JWT Init] Secret initialized - Length: %d, First 5 chars: %s..., Last 5 chars: ...%s", 
			len(secret), 
			secret[:min(5, len(secret))],
			secret[max(0, len(secret)-5):])
		log.Printf("[JWT Init] Secret key bytes length: %d", len(secretKey))
	} else {
		log.Printf("[JWT Init] ERROR: Secret is empty after trimming")
	}
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
	log.Printf("[JWT Verify] Starting token verification...")
	log.Printf("[JWT Verify] Token string length: %d", len(tokenString))
	
	if len(secretKey) == 0 {
		log.Printf("[JWT Verify] ERROR: Secret key is not initialized (length: 0)")
		return nil, errors.New("JWT not initialized")
	}
	
	log.Printf("[JWT Verify] Secret key is initialized (length: %d bytes)", len(secretKey))

	// Trim token string
	tokenString = strings.TrimSpace(tokenString)
	log.Printf("[JWT Verify] Token string after trim: length %d", len(tokenString))

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Log signing method
		method := token.Method.Alg()
		log.Printf("[JWT Verify] Token signing method: %s", method)
		
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("[JWT Verify] ERROR: Invalid signing method - expected HMAC, got: %T", token.Method)
			return nil, fmt.Errorf("invalid signing method: %s (expected HMAC)", method)
		}
		
		log.Printf("[JWT Verify] Signing method is valid (HMAC)")
		log.Printf("[JWT Verify] Using secret key for verification (length: %d bytes)", len(secretKey))
		return secretKey, nil
	})

	if err != nil {
		log.Printf("[JWT Verify] ERROR: Token parsing failed - %v", err)
		log.Printf("[JWT Verify] Error type: %T", err)
		
		// Provide more specific error messages
		if ve, ok := err.(*jwt.ValidationError); ok {
			log.Printf("[JWT Verify] Validation error details - Errors bitmask: %d", ve.Errors)
			
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				log.Printf("[JWT Verify] ERROR: Token is malformed")
				return nil, errors.New("token is malformed")
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				log.Printf("[JWT Verify] ERROR: Token is expired")
				return nil, errors.New("token is expired")
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				log.Printf("[JWT Verify] ERROR: Token is not valid yet")
				return nil, errors.New("token is not valid yet")
			} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				log.Printf("[JWT Verify] ERROR: Signature is invalid - JWT_SECRET mismatch detected")
				log.Printf("[JWT Verify] This usually means JWT_SECRET in filebrowser service doesn't match the main backend")
				return nil, errors.New("signature is invalid - JWT_SECRET may not match the main backend")
			} else if ve.Errors&jwt.ValidationErrorUnverifiable != 0 {
				log.Printf("[JWT Verify] ERROR: Token cannot be verified (signing method mismatch or key issue)")
				return nil, errors.New("token cannot be verified - check signing method and secret key")
			}
			
			log.Printf("[JWT Verify] Other validation error: %v", ve)
		}
		
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		log.Printf("[JWT Verify] SUCCESS: Token is valid")
		log.Printf("[JWT Verify] Claims - UserID: %s, Email: %s, Role: %s", claims.UserID, claims.Email, claims.Role)
		log.Printf("[JWT Verify] Token expiration: %v", claims.ExpiresAt)
		return claims, nil
	}

	log.Printf("[JWT Verify] ERROR: Token claims are invalid or token is not valid")
	return nil, errors.New("invalid token")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

