package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"comment-service/src/conf"
	"comment-service/src/handlers"
)

func main() {
	if err := conf.InitDB(); err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "comment-service"})
	})

	api := r.Group("/api/v1")
	{
		comments := api.Group("/comments")
		{
			comments.GET("/:movieId", handlers.ListComments)
			comments.POST("/:movieId", handlers.CreateComment)
			comments.DELETE("/:id", handlers.DeleteComment)
		}
	}

	log.Printf("comment-service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
