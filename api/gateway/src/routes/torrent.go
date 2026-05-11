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

	movies := r.Group("/api/v1/movies")
	movies.Use(middleware.JWTMiddleware())
	{
		movies.GET("/:id/watched", proxy.ProxyRequest("torrent", "/api/v1/movies/:id/watched"))
		movies.GET("/:id/progress", proxy.ProxyRequest("torrent", "/api/v1/movies/:id/progress"))
		movies.PUT("/:id/progress", proxy.ProxyRequest("torrent", "/api/v1/movies/:id/progress"))
		movies.GET("/:id/subtitles", proxy.ProxyRequest("torrent", "/api/v1/movies/:id/subtitles"))
		movies.GET("/:id/subtitles/:lang", proxy.ProxyRequest("torrent", "/api/v1/movies/:id/subtitles/:lang"))
	}
}
