package pkg

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)

	return string(hash), err
}

func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hash),
		[]byte(password),
	)
}

var (
	hasLower   = regexp.MustCompile(`[a-z]`)
	hasUpper   = regexp.MustCompile(`[A-Z]`)
	hasDigit   = regexp.MustCompile(`\d`)
	hasSpecial = regexp.MustCompile(`[^a-zA-Z\d]`)
)

func ValidatePassword(password string) bool {
	return len(password) >= 8 &&
		hasLower.MatchString(password) &&
		hasUpper.MatchString(password) &&
		hasDigit.MatchString(password) &&
		hasSpecial.MatchString(password)
}
