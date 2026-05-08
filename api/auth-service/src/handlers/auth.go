package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	db "auth-service/src/conf"
	models "auth-service/src/models"
	"auth-service/src/services"
	"auth-service/src/types"
	"auth-service/src/utils"
)

const invalidPayloadPrefix = "invalid payload: "

func CheckAvailabilityHandler(c *gin.Context) {
	var req types.AvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, invalidPayloadPrefix+err.Error())
		return
	}

	if req.Username == "" && req.Email == "" {
		utils.RespondError(c, http.StatusBadRequest, "either username or email must be provided")
		return
	}

	if req.Username != "" {
		available, err := utils.CheckUsernameAvailability(req.Username)
		if err != nil {
			utils.RespondError(c, http.StatusInternalServerError, err.Error())
			return
		}
		if !available {
			suggestions := utils.GenerateUsernameSuggestions(req.Username)
			c.JSON(http.StatusConflict, types.AvailabilityResponse{
				Status:      "error",
				Available:   false,
				Message:     "username déjà utilisé",
				Suggestions: suggestions,
			})
			return
		}
	}

	if req.Email != "" {
		available, err := utils.CheckEmailAvailability(req.Email)
		if err != nil {
			utils.RespondError(c, http.StatusInternalServerError, err.Error())
			return
		}
		if !available {
			c.JSON(http.StatusConflict, types.AvailabilityResponse{
				Status:    "error",
				Available: false,
				Message:   "Email déjà utilisé",
			})
			return
		}
	}

	c.JSON(http.StatusOK, types.AvailabilityResponse{Status: "success", Available: true})
}

func RegisterHandler(c *gin.Context) {
	var req types.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, invalidPayloadPrefix+err.Error())
		return
	}

	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	usernameAvailable, err := utils.CheckUsernameAvailability(req.Username)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !usernameAvailable {
		suggestions := utils.GenerateUsernameSuggestions(req.Username)
		utils.RespondError(c, http.StatusConflict, "Nom d'utilisateur déjà utilisé. Suggestions: "+strings.Join(suggestions, ", "))
		return
	}

	emailAvailable, err := utils.CheckEmailAvailability(req.Email)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	if !emailAvailable {
		utils.RespondError(c, http.StatusConflict, "Email déjà utilisé")
		return
	}

	user, err := services.CreateUser(req)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	if err := sendVerificationCode(user.Email); err != nil {
		log.Printf("failed to queue verification email for %s: %v", user.Email, err)
	}

	tokens, err := utils.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    "Bearer",
		"expires_in":    tokens.ExpiresIn,
	})
}

func LoginHandler(c *gin.Context) {
	var req types.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, invalidPayloadPrefix+err.Error())
		return
	}

	var user models.Users
	if err := db.DB.Where("username = ? OR email = ?", req.Login, req.Login).First(&user).Error; err != nil || user.ID == 0 {
		utils.RespondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "invalid credentials")
		return
	}

	tokens, err := utils.GenerateTokenPair(user.ID, user.Role)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"token_type":    "Bearer",
		"expires_in":    tokens.ExpiresIn,
	})
}

