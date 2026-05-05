package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	db "auth-service/src/conf"
	"auth-service/src/models"
	"auth-service/src/services"
	"auth-service/src/types"
	"auth-service/src/utils"
)

type User = models.Users
type EmailVerification = models.EmailVerification

const emailWhereClause = "email = ?"

func generateVerificationCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, 6)
	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}
	return string(code), nil
}

// sendVerificationCode persists a verification code and emails it.
// Email send failures are non-fatal: registration still succeeds and the code remains valid.
func sendVerificationCode(email string) error {
	code, err := generateVerificationCode()
	if err != nil {
		return fmt.Errorf("failed to generate verification code: %w", err)
	}

	db.DB.Where(emailWhereClause, email).Delete(&EmailVerification{})

	verification := EmailVerification{
		Email:            email,
		VerificationCode: code,
		ExpiresAt:        time.Now().Add(15 * time.Minute),
	}
	if err := db.DB.Create(&verification).Error; err != nil {
		return fmt.Errorf("failed to create verification record: %w", err)
	}

	emailService := services.NewEmailService()
	if err := emailService.SendVerificationEmail(email, code); err != nil {
		fmt.Printf("Failed to send verification email to %s: %v\n", email, err)
		fmt.Printf("Verification code for %s: %s (email failed to send)\n", email, code)
	}

	return nil
}

func SendEmailVerificationHandler(c *gin.Context) {
	// Per-IP: 5 sends per hour.
	ip := c.GetHeader("X-Real-Ip")
	if ip == "" {
		ip = c.RemoteIP()
	}
	if utils.RateLimitRequest("send-verification:ip", ip, 5, time.Hour) {
		utils.RespondError(c, http.StatusTooManyRequests, "Too many requests. Please try again later.")
		return
	}

	var req types.EmailVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Per-email: 3 sends per 15 minutes to prevent mail spam.
	if utils.RateLimitRequest("send-verification:email", req.Email, 3, 15*time.Minute) {
		utils.RespondError(c, http.StatusTooManyRequests, "Too many verification emails sent. Please wait before requesting another.")
		return
	}

	var user User
	if err := db.DB.Where(emailWhereClause, req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RespondError(c, http.StatusNotFound, "User not found")
			return
		}
		utils.RespondError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if user.EmailVerified {
		utils.RespondError(c, http.StatusBadRequest, "Email already verified")
		return
	}

	if err := sendVerificationCode(req.Email); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Verification code sent successfully",
	})
}

func VerifyEmailHandler(c *gin.Context) {
	var req types.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Block after 5 failed attempts within 15 minutes to prevent brute-force
	// of the 6-digit code (1 000 000 combinations).
	const maxAttempts int64 = 5
	if utils.IsRateLimited("verify-email", req.Email, maxAttempts) {
		utils.RespondError(c, http.StatusTooManyRequests, "Too many failed attempts. Please request a new verification code.")
		return
	}

	var verification EmailVerification
	if err := db.DB.Where("email = ? AND verification_code = ?", req.Email, req.VerificationCode).First(&verification).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.RecordFailure("verify-email", req.Email, 15*time.Minute)
			utils.RespondError(c, http.StatusBadRequest, "Invalid verification code")
			return
		}
		utils.RespondError(c, http.StatusInternalServerError, "Database error")
		return
	}

	if time.Now().After(verification.ExpiresAt) {
		db.DB.Delete(&verification)
		utils.RespondError(c, http.StatusBadRequest, "Verification code expired")
		return
	}

	if err := db.DB.Model(&User{}).Where(emailWhereClause, req.Email).Update("email_verified", true).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to update user verification status")
		return
	}

	db.DB.Delete(&verification)
	utils.ClearFailures("verify-email", req.Email)

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}