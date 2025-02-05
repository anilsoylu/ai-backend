package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// CheckPasswordHash şifrenin hash ile eşleşip eşleşmediğini kontrol eder
func CheckPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HashPassword şifreyi hash'ler
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
} 