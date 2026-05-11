package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"torrent-service/src/conf"
	"torrent-service/src/handlers"
	"torrent-service/src/middleware"
	"torrent-service/src/services"
)

func main() {
	conf.InitDB()

	if err := services.InitTorrentClient(); err != nil {
		log.Fatal("Failed to initialize torrent client:", err)
	}

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"status": "ok", "service": "torrent-service"}})
	})

	api := r.Group("/api/v1")
	{
		t := api.Group("/torrent")
		{
			t.POST("/download", handlers.DownloadHandler)
			t.GET("/status/:id", handlers.StatusHandler)
		}
		api.GET("/stream/:id", handlers.StreamHandler)
		api.GET("/movies/:id/watched", handlers.WatchedHandler)
		api.GET("/movies/:id/progress", handlers.GetProgressHandler)
		api.PUT("/movies/:id/progress", handlers.SaveProgressHandler)
		api.GET("/movies/:id/subtitles/:lang", handlers.SubtitleHandler)

		admin := api.Group("/admin")
		admin.Use(middleware.AdminMiddleware())
		{
			admin.GET("/stats", handlers.AdminStatsHandler)
			admin.GET("/films", handlers.AdminListFilmsHandler)
			admin.DELETE("/films/:id", handlers.AdminDeleteFilmHandler)
			admin.POST("/films/:id/download", handlers.AdminReDownloadFilmHandler)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}
	log.Printf("torrent-service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
