package routes

import (
	"gateway/src/middleware"
	"gateway/src/proxy"

	"github.com/gin-gonic/gin"
)

func SetupAdminRoutes(r *gin.Engine) {
	admin := r.Group("/api/v1/admin")
	admin.Use(middleware.JWTMiddleware())
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/users", proxy.ProxyRequest("user", "/api/v1/admin/users"))
		admin.GET("/films", proxy.ProxyRequest("torrent", "/api/v1/admin/films"))
		admin.DELETE("/films/:id", proxy.ProxyRequest("torrent", "/api/v1/admin/films/:id"))
		admin.POST("/films/:id/download", proxy.ProxyRequest("torrent", "/api/v1/admin/films/:id/download"))
	}
}
