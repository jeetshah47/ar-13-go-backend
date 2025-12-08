package password

import (
	"crypto/rand"
	"golang.org/x/crypto/bcrypt"
	"math/big"
)

// HashPassword hashes a password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with a hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateTempPassword generates a secure temporary password
// Length: 10 characters
// Contains: uppercase, lowercase, numbers, and special characters
func GenerateTempPassword() (string, error) {
	const (
		lowercase = "abcdefghijklmnopqrstuvwxyz"
		uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		numbers   = "0123456789"
		special   = "!@#$%^&*"
		allChars  = lowercase + uppercase + numbers + special
		length    = 10
	)

	// Ensure at least one character from each category
	password := make([]byte, length)
	
	// Add one lowercase
	if idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(lowercase)))); err == nil {
		password[0] = lowercase[idx.Int64()]
	} else {
		return "", err
	}
	
	// Add one uppercase
	if idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(uppercase)))); err == nil {
		password[1] = uppercase[idx.Int64()]
	} else {
		return "", err
	}
	
	// Add one number
	if idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(numbers)))); err == nil {
		password[2] = numbers[idx.Int64()]
	} else {
		return "", err
	}
	
	// Add one special character
	if idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(special)))); err == nil {
		password[3] = special[idx.Int64()]
	} else {
		return "", err
	}
	
	// Fill the rest with random characters from all categories
	for i := 4; i < length; i++ {
		if idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(allChars)))); err == nil {
			password[i] = allChars[idx.Int64()]
		} else {
			return "", err
		}
	}
	
	// Shuffle the password to randomize positions
	for i := len(password) - 1; i > 0; i-- {
		if j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1))); err == nil {
			password[i], password[j.Int64()] = password[j.Int64()], password[i]
		} else {
			return "", err
		}
	}
	
	return string(password), nil
}

