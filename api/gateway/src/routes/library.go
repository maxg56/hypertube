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
		library.GET("/movies/:id", proxy.ProxyRequest("library", "/api/v1/library/movies/:id"))
		library.GET("/search", proxy.ProxyRequest("library", "/api/v1/library/search"))
	}
}
