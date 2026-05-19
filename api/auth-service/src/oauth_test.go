package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	models "auth-service/src/models"

	db "auth-service/src/conf"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"encoding/json"
)

func createOAuthTestUser(t *testing.T) models.Users {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("TestPass123!"), bcrypt.DefaultCost)
	require.NoError(t, err)
	user := models.Users{
		Username:     "oauthtestuser",
		Email:        "oauthtest@example.com",
		PasswordHash: string(hash),
		Role:         "user",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	db.DB.Create(&user)
	t.Cleanup(func() { db.DB.Unscoped().Delete(&user) })
	return user
}

func TestOAuthTokenHandler_Password(t *testing.T) {
	router := setupTestRouter()
	createOAuthTestUser(t)

	tests := []struct {
		name           string
		form           url.Values
		expectedStatus int
		expectTokens   bool
		expectErrCode  string
	}{
		{
			name: "valid credentials",
			form: url.Values{
				"grant_type": {"password"},
				"username":   {"oauthtestuser"},
				"password":   {"TestPass123!"},
			},
			expectedStatus: http.StatusOK,
			expectTokens:   true,
		},
		{
			name: "wrong password",
			form: url.Values{
				"grant_type": {"password"},
				"username":   {"oauthtestuser"},
				"password":   {"wrongpassword"},
			},
			expectedStatus: http.StatusUnauthorized,
			expectErrCode:  "invalid_grant",
		},
		{
			name: "unknown user",
			form: url.Values{
				"grant_type": {"password"},
				"username":   {"nobody"},
				"password":   {"TestPass123!"},
			},
			expectedStatus: http.StatusUnauthorized,
			expectErrCode:  "invalid_grant",
		},
		{
			name: "missing username",
			form: url.Values{
				"grant_type": {"password"},
				"password":   {"TestPass123!"},
			},
			expectedStatus: http.StatusBadRequest,
			expectErrCode:  "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/v1/oauth/token", strings.NewReader(tt.form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

			if tt.expectTokens {
				assert.Contains(t, resp, "access_token")
				assert.Contains(t, resp, "refresh_token")
				assert.Equal(t, "Bearer", resp["token_type"])
				assert.Contains(t, resp, "expires_in")
			}
			if tt.expectErrCode != "" {
				assert.Equal(t, tt.expectErrCode, resp["error"])
			}
		})
	}
}

func TestOAuthTokenHandler_RefreshToken(t *testing.T) {
	router := setupTestRouter()
	createOAuthTestUser(t)

	// Obtain a real refresh token via password grant
	form := url.Values{
		"grant_type": {"password"},
		"username":   {"oauthtestuser"},
		"password":   {"TestPass123!"},
	}
	req, _ := http.NewRequest("POST", "/api/v1/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var initial map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &initial))
	validRefresh := initial["refresh_token"].(string)

	tests := []struct {
		name           string
		refreshToken   string
		expectedStatus int
		expectTokens   bool
		expectErrCode  string
	}{
		{
			name:           "valid refresh token",
			refreshToken:   validRefresh,
			expectedStatus: http.StatusOK,
			expectTokens:   true,
		},
		{
			name:           "invalid token",
			refreshToken:   "not.a.valid.token",
			expectedStatus: http.StatusUnauthorized,
			expectErrCode:  "invalid_grant",
		},
		{
			name:           "missing refresh token",
			refreshToken:   "",
			expectedStatus: http.StatusBadRequest,
			expectErrCode:  "invalid_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := url.Values{"grant_type": {"refresh_token"}}
			if tt.refreshToken != "" {
				f.Set("refresh_token", tt.refreshToken)
			}
			req, _ := http.NewRequest("POST", "/api/v1/oauth/token", strings.NewReader(f.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

			if tt.expectTokens {
				assert.Contains(t, resp, "access_token")
				assert.Contains(t, resp, "refresh_token")
				assert.Equal(t, "Bearer", resp["token_type"])
				assert.Contains(t, resp, "expires_in")
			}
			if tt.expectErrCode != "" {
				assert.Equal(t, tt.expectErrCode, resp["error"])
			}
		})
	}
}

func TestOAuthTokenHandler_UnsupportedGrantType(t *testing.T) {
	router := setupTestRouter()

	form := url.Values{"grant_type": {"client_credentials"}}
	req, _ := http.NewRequest("POST", "/api/v1/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "unsupported_grant_type", resp["error"])
}
