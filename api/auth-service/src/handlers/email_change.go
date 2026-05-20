package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	db "auth-service/src/conf"
	"auth-service/src/utils"
)

func getAuthUserID(c *gin.Context) (uint, error) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		utils.RespondError(c, http.StatusUnauthorized, "missing or invalid token")
		return 0, errors.New("unauthorized")
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	claims, err := utils.ParseToken(tokenStr, secret)
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "invalid token")
		return 0, err
	}
	subStr, ok := claims["sub"].(string)
	if !ok || subStr == "" {
		utils.RespondError(c, http.StatusUnauthorized, "invalid token claims")
		return 0, errors.New("invalid claims")
	}
	var userID uint
	if _, err := fmt.Sscanf(subStr, "%d", &userID); err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "invalid user ID")
		return 0, err
	}
	return userID, nil
}

func RequestEmailChangeHandler(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		return
	}

	var req struct {
		NewEmail string `json:"new_email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, invalidPayloadPrefix+err.Error())
		return
	}

	var currentUser User
	if err := db.DB.First(&currentUser, userID).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}
	if currentUser.Email == req.NewEmail {
		utils.RespondError(c, http.StatusBadRequest, "Le nouvel email doit être différent de l'email actuel")
		return
	}

	available, err := utils.CheckEmailAvailability(req.NewEmail)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !available {
		utils.RespondError(c, http.StatusConflict, "Email déjà utilisé")
		return
	}

	if err := sendVerificationCode(req.NewEmail); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Verification code sent to new email",
	})
}

func ConfirmEmailChangeHandler(c *gin.Context) {
	userID, err := getAuthUserID(c)
	if err != nil {
		return
	}

	var req struct {
		NewEmail         string `json:"new_email" binding:"required,email"`
		VerificationCode string `json:"verification_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, invalidPayloadPrefix+err.Error())
		return
	}

	available, err := utils.CheckEmailAvailability(req.NewEmail)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !available {
		utils.RespondError(c, http.StatusConflict, "Email déjà utilisé par un autre compte")
		return
	}

	var verification EmailVerification
	if err := db.DB.Where(
		"email = ? AND verification_code = ? AND expires_at > ?",
		req.NewEmail, req.VerificationCode, time.Now(),
	).First(&verification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondError(c, http.StatusBadRequest, "Code invalide ou expiré")
			return
		}
		utils.RespondError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if err := db.DB.Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"email":          req.NewEmail,
		"email_verified": true,
	}).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update email")
		return
	}

	db.DB.Delete(&verification)

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Email updated successfully",
	})
}
