package services

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"

	db "auth-service/src/conf"
	models "auth-service/src/models"
	"auth-service/src/types"
)

func CreateUser(req types.RegisterRequest) (*models.Users, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to process password")
	}

	user := models.Users{
		Username:  req.Username,
		Email:     req.Email,
		PasswordHash: string(hash),
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &user, nil
}
