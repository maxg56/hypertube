package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerifyTokenHandler(t *testing.T) {
	router := setupTestRouter()

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   "123",
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"exp":   now.Add(15 * time.Minute).Unix(),
		"scope": "access",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	validToken, err := token.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token format",
			authHeader:     "Bearer invalidtoken",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "malformed header",
			authHeader:     "InvalidFormat " + validToken,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/api/v1/auth/verify", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, true, response["success"])
				data := response["data"].(map[string]interface{})
				assert.Equal(t, true, data["valid"])
				assert.Equal(t, "123", data["user_id"])
			} else {
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response, "error")
			}
		})
	}
}

func TestRefreshTokenHandler(t *testing.T) {
	router := setupTestRouter()

	now := time.Now()

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "123",
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"exp":   now.Add(7 * 24 * time.Hour).Unix(),
		"scope": "refresh",
	})
	validRefreshToken, err := refreshToken.SignedString([]byte("test-refresh-secret-key"))
	require.NoError(t, err)

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   "123",
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
		"exp":   now.Add(15 * time.Minute).Unix(),
		"scope": "access",
	})
	validAccessToken, err := accessToken.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	tests := []struct {
		name           string
		payload        map[string]interface{}
		expectedStatus int
	}{
		{
			name:           "valid refresh token",
			payload:        map[string]interface{}{"refresh_token": validRefreshToken},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing refresh token",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid token",
			payload:        map[string]interface{}{"refresh_token": "invalid.token.here"},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "wrong token scope (access token)",
			payload:        map[string]interface{}{"refresh_token": validAccessToken},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonPayload, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(jsonPayload))
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
				assert.Contains(t, data, "expires_in")
			} else {
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response, "error")
			}
		})
	}
}

func TestLogoutHandler(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name           string
		token          string
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "valid token logout",
			token:          "Bearer " + generateTestToken(1),
			expectedStatus: http.StatusOK,
			expectedMsg:    "logged out successfully",
		},
		{
			name:           "missing token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
			expectedMsg:    "missing bearer token",
		},
		{
			name:           "invalid token format",
			token:          "Bearer invalid-jwt",
			expectedStatus: http.StatusOK, // Graceful logout even with invalid token
			expectedMsg:    "logged out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/v1/auth/logout", nil)
			if tt.token != "" {
				req.Header.Set("Authorization", tt.token)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))

			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, true, response["success"])
				data := response["data"].(map[string]interface{})
				assert.Contains(t, data["message"], tt.expectedMsg[:10])
			} else {
				assert.Equal(t, false, response["success"])
				assert.Contains(t, response, "error")
				assert.Equal(t, tt.expectedMsg, response["error"])
			}
		})
	}
}
