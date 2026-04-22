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
		users.GET(profileByID, proxy.ProxyRequest("user", upstreamProfileByID))
		users.GET("/:id/online-status", proxy.ProxyRequest("user", "/api/v1/users/:id/online-status"))

		protected := users.Group("")
		protected.Use(middleware.JWTMiddleware())
		{
			protected.GET("/profile", proxy.ProxyRequest("user", "/api/v1/users/profile"))
			protected.PUT(profileByID, proxy.ProxyRequest("user", upstreamProfileByID))
			protected.DELETE(profileByID, proxy.ProxyRequest("user", upstreamProfileByID))
		}
	}
}
