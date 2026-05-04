package routes

import (
	"gateway/src/middleware"
	"gateway/src/proxy"

	"github.com/gin-gonic/gin"
)

func SetupLibraryRoutes(r *gin.Engine) {
	library := r.Group("/api/v1/library")
	library.Use(middleware.JWTMiddleware())
	{
		library.GET("/movies", proxy.ProxyRequest("library", "/api/v1/library/movies"))
		library.GET("/movies/search", proxy.ProxyRequest("library", "/api/v1/library/movies/search"))
		library.GET("/movies/yts", proxy.ProxyRequest("library", "/api/v1/library/movies/yts"))
		library.GET("/movies/free", proxy.ProxyRequest("library", "/api/v1/library/movies/free"))
		library.GET("/movies/:id", proxy.ProxyRequest("library", "/api/v1/library/movies/:id"))
	}
}
