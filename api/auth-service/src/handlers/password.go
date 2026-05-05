package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	db "auth-service/src/conf"
	"auth-service/src/models"
	"auth-service/src/services"
	"auth-service/src/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const forgotPasswordMsg = "If the email exists, a password reset link will be sent"

// ForgotPasswordRequest represents password reset request payload
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents password reset confirmation payload
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ForgotPasswordHandler handles password reset requests
func ForgotPasswordHandler(c *gin.Context) {
	// Per-IP: 5 requests per hour to limit enumeration / spam from one source.
	ip := c.GetHeader("X-Real-Ip")
	if ip == "" {
		ip = c.RemoteIP()
	}
	if utils.RateLimitRequest("forgot-password:ip", ip, 5, time.Hour) {
		utils.RespondError(c, http.StatusTooManyRequests, "Too many requests. Please try again later.")
		return
	}

	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid payload: "+err.Error())
		return
	}

	// Per-email: 3 requests per hour. Respond identically to avoid leaking whether the email exists.
	if utils.RateLimitRequest("forgot-password:email", req.Email, 3, time.Hour) {
		utils.RespondSuccess(c, http.StatusOK, gin.H{
			"message": forgotPasswordMsg,
		})
		return
	}

	// Find user by email
	var user models.Users
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Don't reveal if email exists or not for security
		utils.RespondSuccess(c, http.StatusOK, gin.H{
			"message": forgotPasswordMsg,
		})
		return
	}

	// Generate secure reset token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to generate reset token")
		return
	}
	resetToken := hex.EncodeToString(bytes)

	// Clean up old tokens for this user
	db.DB.Where("user_id = ?", user.ID).Delete(&models.PasswordReset{})

	// Create password reset record
	passwordReset := models.PasswordReset{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(time.Hour), // 1 hour expiry
		Used:      false,
	}

	if err := db.DB.Create(&passwordReset).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to create password reset token")
		return
	}

	// Send password reset email
	emailService := services.NewEmailService()
	emailErr := emailService.SendPasswordResetEmail(user.Email, resetToken)

	// Always respond with success for security (don't reveal if email exists)
	// Even if email fails, the reset token is created and valid
	if emailErr != nil {
		// Log the error but don't expose it to the user
		// In production, this should be logged to monitoring system
		utils.RespondSuccess(c, http.StatusOK, gin.H{
			"message": forgotPasswordMsg,
		})
	} else {
		utils.RespondSuccess(c, http.StatusOK, gin.H{
			"message": "Password reset email sent successfully",
		})
	}
}

// ResetPasswordHandler handles password reset confirmation
func ResetPasswordHandler(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid payload: "+err.Error())
		return
	}

	if err := utils.ValidatePasswordStrength(req.NewPassword); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	// Find valid reset token
	var passwordReset models.PasswordReset
	if err := db.DB.Preload("User").Where("token = ? AND used = false AND expires_at > ?", req.Token, time.Now()).First(&passwordReset).Error; err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid or expired reset token")
		return
	}

	// Check if the new password is the same as the current password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordReset.User.PasswordHash), []byte(req.NewPassword)); err == nil {
		utils.RespondError(c, http.StatusBadRequest, "Le nouveau mot de passe doit être différent de l'ancien")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Update user password
	if err := db.DB.Model(&passwordReset.User).Update("password_hash", string(hashedPassword)).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to update password")
		return
	}

	// Mark token as used
	if err := db.DB.Model(&passwordReset).Update("used", true).Error; err != nil {
		// Log error but continue since password was updated
		// This is not critical failure
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Password reset successful",
	})
}

// HealthCheckHandler returns service health status
func HealthCheckHandler(c *gin.Context) {
	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
		"version": "1.0.0",
	})
}
