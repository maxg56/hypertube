package main

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	db "auth-service/src/conf"
	"auth-service/src/handlers"
	models "auth-service/src/models"
)

func setupTestDB() *gorm.DB {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}
	database.AutoMigrate(&models.Users{}, &models.PasswordReset{})
	return database
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/health", handlers.HealthCheckHandler)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.RegisterHandler)
			auth.POST("/login", handlers.LoginHandler)
			auth.POST("/logout", handlers.LogoutHandler)
			auth.POST("/refresh", handlers.RefreshTokenHandler)
			auth.GET("/verify", handlers.VerifyTokenHandler)
			auth.POST("/forgot-password", handlers.ForgotPasswordHandler)
			auth.POST("/reset-password", handlers.ResetPasswordHandler)
		}
	}

	return r
}

func generateTestToken(userID int) string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "test-secret-key"
	}

	claims := jwt.MapClaims{
		"sub":     "1",
		"user_id": userID,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestMain(m *testing.M) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret-key")
	os.Setenv("JWT_ACCESS_TTL", "15m")
	os.Setenv("JWT_REFRESH_TTL", "7d")

	db.DB = setupTestDB()

	os.Exit(m.Run())
}
