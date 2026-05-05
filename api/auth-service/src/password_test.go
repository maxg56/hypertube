package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "auth-service/src/conf"
	models "auth-service/src/models"
)

func TestForgotPasswordHandler(t *testing.T) {
	db.DB = setupTestDB()
	router := setupTestRouter()

	tests := []struct {
		name       string
		payload    map[string]interface{}
		statusCode int
		setupUser  bool
	}{
		{
			name:       "valid email with existing user",
			payload:    map[string]interface{}{"email": "test@example.com"},
			statusCode: http.StatusOK,
			setupUser:  true,
		},
		{
			name:       "valid email with non-existing user",
			payload:    map[string]interface{}{"email": "nonexistent@example.com"},
			statusCode: http.StatusOK, // returns OK for security
			setupUser:  false,
		},
		{
			name:       "invalid email format",
			payload:    map[string]interface{}{"email": "invalid-email"},
			statusCode: http.StatusBadRequest,
			setupUser:  false,
		},
		{
			name:       "missing email",
			payload:    map[string]interface{}{},
			statusCode: http.StatusBadRequest,
			setupUser:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupUser {
				user := models.Users{
					Username:     "testuser",
					FirstName:    "Test",
					LastName:     "User",
					Email:        tt.payload["email"].(string),
					PasswordHash: "$2a$10$abcdefg",
				}
				db.DB.Create(&user)
				defer db.DB.Delete(&user)
			}

			jsonBytes, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/v1/auth/forgot-password", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			var response map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

			if tt.statusCode == http.StatusOK {
				assert.Equal(t, true, response["success"])
				data := response["data"].(map[string]interface{})
				message := data["message"].(string)
				assert.True(t, strings.Contains(message, "password reset") || strings.Contains(message, "Password reset"))
			} else {
				assert.Equal(t, false, response["success"])
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

func TestResetPasswordHandler(t *testing.T) {
	db.DB = setupTestDB()
	router := setupTestRouter()

	user := models.Users{
		Username:     "testuser",
		FirstName:    "Test",
		LastName:     "User",
		Email:        "test@example.com",
		PasswordHash: "$2a$10$abcdefg",
	}
	db.DB.Create(&user)
	defer db.DB.Delete(&user)

	validToken := models.PasswordReset{
		UserID:    user.ID,
		Token:     "valid-token-123",
		ExpiresAt: time.Now().Add(time.Hour),
		Used:      false,
	}
	db.DB.Create(&validToken)
	defer db.DB.Delete(&validToken)

	expiredToken := models.PasswordReset{
		UserID:    user.ID,
		Token:     "expired-token-123",
		ExpiresAt: time.Now().Add(-time.Hour),
		Used:      false,
	}
	db.DB.Create(&expiredToken)
	defer db.DB.Delete(&expiredToken)

	usedToken := models.PasswordReset{
		UserID:    user.ID,
		Token:     "used-token-123",
		ExpiresAt: time.Now().Add(time.Hour),
		Used:      true,
	}
	db.DB.Create(&usedToken)
	defer db.DB.Delete(&usedToken)

	tests := []struct {
		name       string
		payload    map[string]interface{}
		statusCode int
	}{
		{
			name:       "valid token and password",
			payload:    map[string]interface{}{"token": "valid-token-123", "new_password": "newpassword123"},
			statusCode: http.StatusOK,
		},
		{
			name:       "expired token",
			payload:    map[string]interface{}{"token": "expired-token-123", "new_password": "newpassword123"},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "used token",
			payload:    map[string]interface{}{"token": "used-token-123", "new_password": "newpassword123"},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "invalid token",
			payload:    map[string]interface{}{"token": "invalid-token", "new_password": "newpassword123"},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "missing token",
			payload:    map[string]interface{}{"new_password": "newpassword123"},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "short password",
			payload:    map[string]interface{}{"token": "valid-token-123", "new_password": "short"},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "same password as current",
			payload:    map[string]interface{}{"token": "valid-token-123", "new_password": "password"},
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/v1/auth/reset-password", bytes.NewBuffer(jsonBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)

			var response map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

			if tt.statusCode == http.StatusOK {
				assert.Equal(t, true, response["success"])
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data["message"], "Password reset successful")

				var updated models.PasswordReset
				db.DB.Where("token = ?", tt.payload["token"].(string)).First(&updated)
				assert.Equal(t, true, updated.Used)
			} else {
				assert.Equal(t, false, response["success"])
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}
