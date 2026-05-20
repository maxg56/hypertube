package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	db "auth-service/src/conf"
	models "auth-service/src/models"
	"auth-service/src/utils"
)

const oauthStateTTL = 5 * time.Minute

var nonAlphanumRegex = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

func generateOAuthState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func storeOAuthState(state, callbackURL string) error {
	if db.Client == nil {
		return fmt.Errorf("redis not available")
	}
	return db.Client.Set(context.Background(), "oauth_state:"+state, callbackURL, oauthStateTTL).Err()
}

func consumeOAuthState(state string) (string, error) {
	if db.Client == nil {
		return "/", nil
	}
	ctx := context.Background()
	val, err := db.Client.GetDel(ctx, "oauth_state:"+state).Result()
	if err != nil {
		return "", fmt.Errorf("état OAuth invalide ou expiré")
	}
	return val, nil
}

func sanitizeUsername(raw string) string {
	clean := nonAlphanumRegex.ReplaceAllString(raw, "_")
	if len(clean) > 30 {
		clean = clean[:30]
	}
	for len(clean) < 3 {
		clean += "_"
	}
	return clean
}

func generateUniqueUsername(base string) string {
	base = sanitizeUsername(base)
	available, err := utils.CheckUsernameAvailability(base)
	if err == nil && available {
		return base
	}
	suggestions := utils.GenerateUsernameSuggestions(base)
	if len(suggestions) > 0 {
		return suggestions[0]
	}
	b := make([]byte, 4)
	rand.Read(b)
	return base[:min(20, len(base))] + hex.EncodeToString(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// findOrCreateOAuthUser looks up or creates a user for the given OAuth identity.
func findOrCreateOAuthUser(provider, providerID, email, firstName, lastName, avatarURL, login string) (*models.Users, error) {
	var oauthAcc models.OAuthAccount
	err := db.DB.Where("provider = ? AND provider_id = ?", provider, providerID).First(&oauthAcc).Error
	if err == nil {
		var user models.Users
		if err := db.DB.First(&user, oauthAcc.UserID).Error; err != nil {
			return nil, fmt.Errorf("user not found for oauth account: %w", err)
		}
		return &user, nil
	}

	// Link to existing account if email matches
	if email != "" {
		var existing models.Users
		if err := db.DB.Where("email = ?", email).First(&existing).Error; err == nil {
			link := models.OAuthAccount{
				UserID:     existing.ID,
				Provider:   provider,
				ProviderID: providerID,
			}
			if err := db.DB.Create(&link).Error; err != nil {
				return nil, fmt.Errorf("failed to link OAuth account: %w", err)
			}
			return &existing, nil
		}
	}

	// Create new user
	username := generateUniqueUsername(login)
	if email == "" {
		email = username + "@oauth.placeholder"
	}

	user := models.Users{
		Username:      username,
		Email:         email,
		PasswordHash:  "OAUTH_NO_PASSWORD",
		FirstName:     firstName,
		LastName:      lastName,
		AvatarURL:     avatarURL,
		EmailVerified: true,
		Role:          models.RoleUser,
	}
	if err := db.DB.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	link := models.OAuthAccount{
		UserID:     user.ID,
		Provider:   provider,
		ProviderID: providerID,
	}
	if err := db.DB.Create(&link).Error; err != nil {
		return nil, fmt.Errorf("failed to create OAuth link: %w", err)
	}

	return &user, nil
}

func redirectWithError(c *gin.Context, frontendURL, msg string) {
	dest := frontendURL + "/login?error=" + url.QueryEscape(msg)
	c.Redirect(http.StatusTemporaryRedirect, dest)
}

func redirectWithTokens(c *gin.Context, frontendURL, callbackURL string, tokens *utils.TokenPair) {
	if !strings.HasPrefix(callbackURL, "/") {
		callbackURL = "/"
	}
	params := url.Values{}
	params.Set("access_token", tokens.AccessToken)
	params.Set("refresh_token", tokens.RefreshToken)
	params.Set("expires_in", fmt.Sprintf("%d", tokens.ExpiresIn))
	params.Set("callbackUrl", callbackURL)
	c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/oauth/callback?"+params.Encode())
}

// --- 42 School ---

// OAuth42RedirectHandler handles GET /auth/oauth/42
func OAuth42RedirectHandler(c *gin.Context) {
	clientID := os.Getenv("OAUTH_42_CLIENT_ID")
	redirectURI := os.Getenv("OAUTH_42_REDIRECT_URI")
	frontendURL := db.GetenvOrDefault("FRONTEND_URL", "http://localhost:8443")

	if clientID == "" || redirectURI == "" {
		redirectWithError(c, frontendURL, "42 OAuth non configuré")
		return
	}

	state, err := generateOAuthState()
	if err != nil {
		redirectWithError(c, frontendURL, "erreur interne")
		return
	}

	callbackURL := c.Query("callbackUrl")
	if callbackURL == "" {
		callbackURL = "/"
	}
	if err := storeOAuthState(state, callbackURL); err != nil {
		redirectWithError(c, frontendURL, "erreur interne")
		return
	}

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "public")
	params.Set("state", state)

	c.Redirect(http.StatusTemporaryRedirect, "https://api.intra.42.fr/oauth/authorize?"+params.Encode())
}

// OAuth42CallbackHandler handles GET /auth/oauth/42/callback
func OAuth42CallbackHandler(c *gin.Context) {
	frontendURL := db.GetenvOrDefault("FRONTEND_URL", "http://localhost:8443")

	if errMsg := c.Query("error"); errMsg != "" {
		redirectWithError(c, frontendURL, errMsg)
		return
	}

	state := c.Query("state")
	callbackURL, err := consumeOAuthState(state)
	if err != nil {
		redirectWithError(c, frontendURL, "état OAuth invalide")
		return
	}

	code := c.Query("code")
	if code == "" {
		redirectWithError(c, frontendURL, "code manquant")
		return
	}

	accessToken, err := exchange42Token(code)
	if err != nil {
		redirectWithError(c, frontendURL, "échange de token échoué")
		return
	}

	info, err := fetch42UserInfo(accessToken)
	if err != nil {
		redirectWithError(c, frontendURL, "récupération du profil échouée")
		return
	}

	user, err := findOrCreateOAuthUser("42", info.providerID, info.email, info.firstName, info.lastName, info.avatarURL, info.login)
	if err != nil {
		redirectWithError(c, frontendURL, "création du compte échouée")
		return
	}

	tokens, err := utils.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		redirectWithError(c, frontendURL, "erreur interne")
		return
	}

	redirectWithTokens(c, frontendURL, callbackURL, tokens)
}

type oauthUserInfo struct {
	providerID string
	login      string
	email      string
	firstName  string
	lastName   string
	avatarURL  string
}

func exchange42Token(code string) (string, error) {
	clientID := os.Getenv("OAUTH_42_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_42_CLIENT_SECRET")
	redirectURI := os.Getenv("OAUTH_42_REDIRECT_URI")

	resp, err := http.PostForm("https://api.intra.42.fr/oauth/token", url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
		"redirect_uri":  {redirectURI},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token exchange failed: %d", resp.StatusCode)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}
	return result.AccessToken, nil
}

func fetch42UserInfo(accessToken string) (*oauthUserInfo, error) {
	req, err := http.NewRequest("GET", "https://api.intra.42.fr/v2/me", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Image     struct {
			Link string `json:"link"`
		} `json:"image"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	return &oauthUserInfo{
		providerID: fmt.Sprintf("%d", raw.ID),
		login:      raw.Login,
		email:      raw.Email,
		firstName:  raw.FirstName,
		lastName:   raw.LastName,
		avatarURL:  raw.Image.Link,
	}, nil
}

// --- GitHub ---

// OAuthGithubRedirectHandler handles GET /auth/oauth/github
func OAuthGithubRedirectHandler(c *gin.Context) {
	clientID := os.Getenv("OAUTH_GITHUB_CLIENT_ID")
	frontendURL := db.GetenvOrDefault("FRONTEND_URL", "http://localhost:8443")

	if clientID == "" {
		redirectWithError(c, frontendURL, "GitHub OAuth non configuré")
		return
	}

	state, err := generateOAuthState()
	if err != nil {
		redirectWithError(c, frontendURL, "erreur interne")
		return
	}

	callbackURL := c.Query("callbackUrl")
	if callbackURL == "" {
		callbackURL = "/"
	}
	if err := storeOAuthState(state, callbackURL); err != nil {
		redirectWithError(c, frontendURL, "erreur interne")
		return
	}

	redirectURI := os.Getenv("OAUTH_GITHUB_REDIRECT_URI")

	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("scope", "user:email")
	params.Set("state", state)
	if redirectURI != "" {
		params.Set("redirect_uri", redirectURI)
	}

	c.Redirect(http.StatusTemporaryRedirect, "https://github.com/login/oauth/authorize?"+params.Encode())
}

// OAuthGithubCallbackHandler handles GET /auth/oauth/github/callback
func OAuthGithubCallbackHandler(c *gin.Context) {
	frontendURL := db.GetenvOrDefault("FRONTEND_URL", "http://localhost:8443")

	if errMsg := c.Query("error"); errMsg != "" {
		redirectWithError(c, frontendURL, errMsg)
		return
	}

	state := c.Query("state")
	callbackURL, err := consumeOAuthState(state)
	if err != nil {
		redirectWithError(c, frontendURL, "état OAuth invalide")
		return
	}

	code := c.Query("code")
	if code == "" {
		redirectWithError(c, frontendURL, "code manquant")
		return
	}

	accessToken, err := exchangeGithubToken(code)
	if err != nil {
		redirectWithError(c, frontendURL, "échange de token échoué")
		return
	}

	info, err := fetchGithubUserInfo(accessToken)
	if err != nil {
		redirectWithError(c, frontendURL, "récupération du profil échouée")
		return
	}

	user, err := findOrCreateOAuthUser("github", info.providerID, info.email, info.firstName, info.lastName, info.avatarURL, info.login)
	if err != nil {
		redirectWithError(c, frontendURL, "création du compte échouée")
		return
	}

	tokens, err := utils.GenerateTokenPair(user.ID, string(user.Role))
	if err != nil {
		redirectWithError(c, frontendURL, "erreur interne")
		return
	}

	redirectWithTokens(c, frontendURL, callbackURL, tokens)
}

func exchangeGithubToken(code string) (string, error) {
	clientID := os.Getenv("OAUTH_GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_GITHUB_CLIENT_SECRET")
	redirectURI := os.Getenv("OAUTH_GITHUB_REDIRECT_URI")

	params := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}
	if redirectURI != "" {
		params.Set("redirect_uri", redirectURI)
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(params.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		AccessToken string `json:"access_token"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Error != "" {
		return "", fmt.Errorf("github token error: %s", result.Error)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token")
	}
	return result.AccessToken, nil
}

func fetchGithubUserInfo(accessToken string) (*oauthUserInfo, error) {
	doReq := func(urlStr string) ([]byte, error) {
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

		c := &http.Client{Timeout: 10 * time.Second}
		resp, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return io.ReadAll(resp.Body)
	}

	body, err := doReq("https://api.github.com/user")
	if err != nil {
		return nil, err
	}

	var raw struct {
		ID        int    `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	email := raw.Email
	if email == "" {
		emailBody, err := doReq("https://api.github.com/user/emails")
		if err == nil {
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.Unmarshal(emailBody, &emails) == nil {
				for _, e := range emails {
					if e.Primary {
						email = e.Email
						break
					}
				}
				if email == "" && len(emails) > 0 {
					email = emails[0].Email
				}
			}
		}
	}

	firstName, lastName := raw.Login, ""
	if raw.Name != "" {
		parts := strings.SplitN(raw.Name, " ", 2)
		firstName = parts[0]
		if len(parts) == 2 {
			lastName = parts[1]
		}
	}

	return &oauthUserInfo{
		providerID: fmt.Sprintf("%d", raw.ID),
		login:      raw.Login,
		email:      email,
		firstName:  firstName,
		lastName:   lastName,
		avatarURL:  raw.AvatarURL,
	}, nil
}
