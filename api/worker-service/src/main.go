package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"

	"worker-service/src/conf"
	"worker-service/src/handlers"
	"worker-service/src/services"
)

func main() {
	conf.InitDB()

	c := cron.New()
	if _, err := c.AddFunc("*/10 * * * *", services.CleanupUnwatchedMovies); err != nil {
		log.Fatalf("[MAIN] erreur configuration cron: %v", err)
	}
	c.Start()
	defer c.Stop()
	log.Println("[MAIN] cron démarré")

	go services.CleanupUnwatchedMovies()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8006"
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "worker-service"})
	})

	r.POST("/delete-torrent", handlers.DeleteTorrentHandler)

	log.Printf("[MAIN] worker-service starting on port %s", port)
	log.Fatal(r.Run(":" + port))
}
