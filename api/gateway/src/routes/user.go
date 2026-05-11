package routes

import (
	"gateway/src/middleware"
	"gateway/src/proxy"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(r *gin.Engine) {
	const (
		profileByID         = "/profile/:id"
		upstreamProfileByID = "/api/v1/users/profile/:id"
	)

	users := r.Group("/api/v1/users")
	{
		users.GET("/avatars/*filename", proxy.ProxyRequest("user", "/api/v1/users/avatars*filename"))
		users.GET("/search", proxy.ProxyRequest("user", "/api/v1/users/search"))
		users.GET(profileByID, proxy.ProxyRequest("user", upstreamProfileByID))
		users.GET("/:id/online-status", proxy.ProxyRequest("user", "/api/v1/users/:id/online-status"))
		users.GET("/:id/favorites", proxy.ProxyRequest("user", "/api/v1/users/:id/favorites"))

		protected := users.Group("")
		protected.Use(middleware.JWTMiddleware())
		{
			protected.GET("", proxy.ProxyRequest("user", "/api/v1/users"))
			protected.GET("/profile", proxy.ProxyRequest("user", "/api/v1/users/profile"))
			protected.PUT(profileByID, proxy.ProxyRequest("user", upstreamProfileByID))
			protected.DELETE(profileByID, proxy.ProxyRequest("user", upstreamProfileByID))
			protected.POST("/avatar", proxy.ProxyRequest("user", "/api/v1/users/avatar"))

			protected.GET("/favorites", proxy.ProxyRequest("user", "/api/v1/users/favorites"))
			protected.POST("/favorites", proxy.ProxyRequest("user", "/api/v1/users/favorites"))
			protected.DELETE("/favorites/:tmdbId", proxy.ProxyRequest("user", "/api/v1/users/favorites/:tmdbId"))
			protected.GET("/favorites/:tmdbId", proxy.ProxyRequest("user", "/api/v1/users/favorites/:tmdbId"))

			protected.GET("/watch-later", proxy.ProxyRequest("user", "/api/v1/users/watch-later"))
			protected.POST("/watch-later", proxy.ProxyRequest("user", "/api/v1/users/watch-later"))
			protected.DELETE("/watch-later/:tmdbId", proxy.ProxyRequest("user", "/api/v1/users/watch-later/:tmdbId"))
			protected.GET("/watch-later/:tmdbId", proxy.ProxyRequest("user", "/api/v1/users/watch-later/:tmdbId"))
		}
	}
}
