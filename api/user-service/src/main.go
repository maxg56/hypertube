package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/handlers"
	"user-service/src/middleware"
)

func main() {
	conf.InitDB()

	if err := conf.InitRedis(); err != nil {
		log.Printf("Warning: Redis unavailable: %v", err)
	} else {
		log.Println("Redis connected")
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "user-service"})
	})

	avatarDir := os.Getenv("AVATAR_DIR")
	if avatarDir == "" {
		avatarDir = "/data/avatars"
	}
	r.Static("/api/v1/users/avatars", avatarDir)

	const profileByID = "/profile/:id"

	users := r.Group("/api/v1/users")
	{
		users.GET("/search", handlers.SearchUsersHandler)
		users.GET(profileByID, handlers.GetProfileHandler)
		users.GET("/:id/online-status", handlers.GetUserOnlineStatusHandler)
		users.GET("/:id/favorites", handlers.ListUserFavoritesHandler)

		protected := users.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("", handlers.ListUsersHandler)
			protected.GET("/profile", handlers.GetOwnProfileHandler)
			protected.PUT(profileByID, handlers.UpdateProfileHandler)
			protected.DELETE(profileByID, handlers.DeleteProfileHandler)
			protected.POST("/avatar", handlers.UploadAvatarHandler)

			protected.GET("/favorites", handlers.ListFavoritesHandler)
			protected.POST("/favorites", handlers.AddFavoriteHandler)
			protected.DELETE("/favorites/:tmdbId", handlers.RemoveFavoriteHandler)
			protected.GET("/favorites/:tmdbId", handlers.CheckFavoriteHandler)

			protected.GET("/watch-later", handlers.ListWatchLaterHandler)
			protected.POST("/watch-later", handlers.AddWatchLaterHandler)
			protected.DELETE("/watch-later/:tmdbId", handlers.RemoveWatchLaterHandler)
			protected.GET("/watch-later/:tmdbId", handlers.CheckWatchLaterHandler)
		}
	}

	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware())
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/users", handlers.AdminListUsersHandler)
		admin.PUT("/users/:id/role", handlers.AdminPromoteUserHandler)
		admin.DELETE("/users/:id", handlers.AdminDeleteUserHandler)
		admin.PUT("/users/:id/username", handlers.AdminUpdateUsernameHandler)
	}

	log.Println("User service starting on port 8002")
	log.Fatal(http.ListenAndServe(":8002", r))
}
