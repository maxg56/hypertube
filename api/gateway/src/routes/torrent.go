package routes

import (
	"gateway/src/middleware"
	"gateway/src/proxy"

	"github.com/gin-gonic/gin"
)

func SetupTorrentRoutes(r *gin.Engine) {
	torrent := r.Group("/api/v1/torrent")
	torrent.Use(middleware.JWTMiddleware())
	{
		torrent.POST("/download", proxy.ProxyRequest("torrent", "/api/v1/torrent/download"))
		torrent.GET("/status/:id", proxy.ProxyRequest("torrent", "/api/v1/torrent/status/:id"))
	}

	stream := r.Group("/api/v1/stream")
	stream.Use(middleware.JWTMiddleware())
	{
		stream.GET("/:id", proxy.ProxyRequest("torrent", "/api/v1/stream/:id"))
	}

	subtitle := r.Group("/api/v1/subtitle")
	subtitle.Use(middleware.JWTMiddleware())
	{
		subtitle.GET("/:id", proxy.ProxyRequest("torrent", "/api/v1/subtitle/:id"))
	}
}
