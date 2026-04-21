package main

import (
	"log"
	"net/http"

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

	const profileByID = "/profile/:id"

	users := r.Group("/api/v1/users")
	{
		users.GET(profileByID, handlers.GetProfileHandler)
		users.GET("/:id/online-status", handlers.GetUserOnlineStatusHandler)

		protected := users.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.GET("/profile", handlers.GetOwnProfileHandler)
			protected.PUT(profileByID, handlers.UpdateProfileHandler)
			protected.DELETE(profileByID, handlers.DeleteProfileHandler)
		}
	}

	log.Println("User service starting on port 8002")
	log.Fatal(http.ListenAndServe(":8002", r))
}
