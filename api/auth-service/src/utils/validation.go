package utils

import (
	"errors"
	"fmt"
	"unicode"

	db "auth-service/src/conf"
	models "auth-service/src/models"
	"gorm.io/gorm"
)

// ValidatePasswordStrength enforces complexity rules beyond the minimum length.
func ValidatePasswordStrength(password string) error {
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}
	return nil
}

// CheckUsernameAvailability checks if a username is available
func CheckUsernameAvailability(username string) (bool, error) {
	var user models.Users
	err := db.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, fmt.Errorf("database error: %w", err)
	}
	return false, nil
}

// CheckEmailAvailability checks if an email is available
func CheckEmailAvailability(email string) (bool, error) {
	var user models.Users
	err := db.DB.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, fmt.Errorf("database error: %w", err)
	}
	return false, nil
}
