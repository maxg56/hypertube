package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	db "auth-service/src/conf"
	"auth-service/src/handlers"
	"auth-service/src/utils"
)

func main() {
	// Initialize database (will AutoMigrate models)
	db.ConnectDatabase()

	// Initialize Redis for token blacklisting and rate limiting
	if err := db.InitRedis(); err != nil {
		log.Printf("Failed to initialize Redis: %v", err)
		log.Println("Redis initialization failed - tokens will not be blacklisted on logout")
	} else {
		log.Println("Redis initialized successfully for token blacklisting")
		utils.InitRateLimiter()
	}

	r := gin.Default()

	// Health check
	r.GET("/health", handlers.HealthCheckHandler)

	// API routes
	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/check-availability", handlers.CheckAvailabilityHandler)
			auth.POST("/register", handlers.RegisterHandler)
			auth.POST("/login", handlers.LoginHandler)
			auth.POST("/logout", handlers.LogoutHandler)
			auth.POST("/refresh", handlers.RefreshTokenHandler)
			auth.GET("/verify", handlers.VerifyTokenHandler)
			auth.POST("/forgot-password", handlers.ForgotPasswordHandler)
			auth.POST("/reset-password", handlers.ResetPasswordHandler)
			auth.POST("/send-email-verification", handlers.SendEmailVerificationHandler)
			auth.POST("/verify-email", handlers.VerifyEmailHandler)
		}

		// OAuth2 token endpoint (RFC 6749)
		oauth := api.Group("/oauth")
		{
			oauth.POST("/token", handlers.OAuthTokenHandler)
		}

		// Social OAuth2 (42 School, GitHub)
		social := auth.Group("/oauth")
		{
			social.GET("/42", handlers.OAuth42RedirectHandler)
			social.GET("/42/callback", handlers.OAuth42CallbackHandler)
			social.GET("/github", handlers.OAuthGithubRedirectHandler)
			social.GET("/github/callback", handlers.OAuthGithubCallbackHandler)
		}
	}

	log.Println("Auth service starting on port 8001")
	log.Fatal(http.ListenAndServe(":8001", r))
}
