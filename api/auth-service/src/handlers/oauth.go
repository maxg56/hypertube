package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	db "auth-service/src/conf"
	models "auth-service/src/models"
	"auth-service/src/utils"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// oauthError writes an RFC 6749 error response.
func oauthError(c *gin.Context, status int, code, description string) {
	c.JSON(status, gin.H{
		"error":             code,
		"error_description": description,
	})
}

// OAuthTokenHandler implements POST /oauth/token (RFC 6749).
// Accepts application/x-www-form-urlencoded or application/json.
func OAuthTokenHandler(c *gin.Context) {
	grantType := c.PostForm("grant_type")
	if grantType == "" {
		// Fallback: try JSON body
		var body struct {
			GrantType    string `json:"grant_type"`
			Username     string `json:"username"`
			Password     string `json:"password"`
			RefreshToken string `json:"refresh_token"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			oauthError(c, http.StatusBadRequest, "invalid_request", "grant_type is required")
			return
		}
		grantType = body.GrantType
	}

	switch grantType {
	case "password":
		oauthPassword(c)
	case "refresh_token":
		oauthRefresh(c)
	default:
		oauthError(c, http.StatusBadRequest, "unsupported_grant_type", "supported grant types: password, refresh_token")
	}
}

func oauthPassword(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	if username == "" || password == "" {
		oauthError(c, http.StatusBadRequest, "invalid_request", "username and password are required")
		return
	}

	var user models.Users
	if err := db.DB.Where("username = ? OR email = ?", username, username).First(&user).Error; err != nil || user.ID == 0 {
		oauthError(c, http.StatusUnauthorized, "invalid_grant", "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		oauthError(c, http.StatusUnauthorized, "invalid_grant", "invalid credentials")
		return
	}

	tokens, err := utils.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to issue token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":   tokens.AccessToken,
		"refresh_token":  tokens.RefreshToken,
		"token_type":     "Bearer",
		"expires_in":     tokens.ExpiresIn,
		"email_verified": user.EmailVerified,
		"email":          user.Email,
	})
}

func oauthRefresh(c *gin.Context) {
	refreshToken := c.PostForm("refresh_token")
	if refreshToken == "" {
		oauthError(c, http.StatusBadRequest, "invalid_request", "refresh_token is required")
		return
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		oauthError(c, http.StatusInternalServerError, "server_error", "server misconfigured")
		return
	}
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = secret
	}

	claims, err := utils.ParseToken(refreshToken, refreshSecret)
	if err != nil {
		oauthError(c, http.StatusUnauthorized, "invalid_grant", "invalid refresh token")
		return
	}

	if blacklisted, err := db.IsTokenBlacklisted(refreshToken); err == nil && blacklisted {
		oauthError(c, http.StatusUnauthorized, "invalid_grant", "refresh token has been revoked")
		return
	}

	scope, ok := claims["scope"].(string)
	if !ok || scope != "refresh" {
		oauthError(c, http.StatusUnauthorized, "invalid_grant", "invalid token scope")
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		oauthError(c, http.StatusUnauthorized, "invalid_grant", "invalid token claims")
		return
	}

	var userIDUint uint
	if _, err := fmt.Sscanf(userID, "%d", &userIDUint); err != nil {
		oauthError(c, http.StatusUnauthorized, "invalid_grant", "invalid user ID format")
		return
	}

	role := "user"
	if r, ok := claims["role"].(string); ok && r != "" {
		role = r
	}

	accessTTL := utils.GetDurationFromEnv("JWT_ACCESS_TTL", 15*time.Minute)
	refreshTTL := utils.GetDurationFromEnv("JWT_REFRESH_TTL", 7*24*time.Hour)

	now := time.Now()
	accessToken, err := utils.SignToken(jwt.MapClaims{
		"sub":   userID,
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"exp":   now.Add(accessTTL).Unix(),
		"scope": "access",
		"role":  role,
	}, secret)
	if err != nil {
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to issue token")
		return
	}

	newRefreshToken, err := utils.SignToken(jwt.MapClaims{
		"sub":   userID,
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"exp":   now.Add(refreshTTL).Unix(),
		"scope": "refresh",
		"role":  role,
	}, refreshSecret)
	if err != nil {
		oauthError(c, http.StatusInternalServerError, "server_error", "failed to issue token")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
		"token_type":    "Bearer",
		"expires_in":    int64(accessTTL.Seconds()),
	})
}
