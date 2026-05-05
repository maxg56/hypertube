package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "auth-service/src/conf"
	models "auth-service/src/models"
)

func TestRegisterHandler(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid registration",
			payload: map[string]interface{}{
				"username":          "testuser",
				"email":             "test@example.com",
				"password":          "password123",
				"first_name":        "Test",
				"last_name":         "User",
				"birth_date":        "1990-01-15",
				"gender":            "man",
				"sex_pref":          "both",
				"relationship_type": "long_term",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			payload: map[string]interface{}{
				"username": "testuser",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate username",
			payload: map[string]interface{}{
				"username":          "testuser",
				"email":             "different@example.com",
				"password":          "password123",
				"first_name":        "Another",
				"last_name":         "User",
				"birth_date":        "1985-05-20",
				"gender":            "woman",
				"sex_pref":          "man",
				"relationship_type": "short_term",
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

			if tt.expectedStatus == http.StatusCreated {
				assert.Equal(t, true, response["success"])
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "access_token")
				assert.Contains(t, data, "refresh_token")
				assert.Equal(t, "Bearer", data["token_type"])
				assert.Contains(t, data, "expires_in")
			} else {
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response, "error")
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	router := setupTestRouter()

	user := models.Users{
		Username:     "loginuser",
		Email:        "login@example.com",
		PasswordHash: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // "password"
		FirstName:    "Login",
		LastName:     "User",
	}
	db.DB.Create(&user)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid login with username",
			payload: map[string]interface{}{
				"login":    "loginuser",
				"password": "password",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "valid login with email",
			payload: map[string]interface{}{
				"login":    "login@example.com",
				"password": "password",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid password",
			payload: map[string]interface{}{
				"login":    "loginuser",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "nonexistent user",
			payload: map[string]interface{}{
				"login":    "nonexistent",
				"password": "password",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, true, response["success"])
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data, "access_token")
				assert.Contains(t, data, "refresh_token")
				assert.Equal(t, "Bearer", data["token_type"])
			} else {
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response, "error")
			}
		})
	}
}
