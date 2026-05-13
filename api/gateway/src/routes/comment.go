package routes

import (
	"gateway/src/middleware"
	"gateway/src/proxy"

	"github.com/gin-gonic/gin"
)

func SetupCommentRoutes(r *gin.Engine) {
	comment := r.Group("/api/v1/comments")
	comment.Use(middleware.JWTMiddleware())
	{
		comment.GET("/user/:userId", proxy.ProxyRequest("comment", "/api/v1/comments/user/:userId"))
		comment.GET("/:movieId", proxy.ProxyRequest("comment", "/api/v1/comments/:movieId"))

		write := comment.Group("")
		write.Use(middleware.RequireEmailVerified())
		{
			write.POST("/:movieId", proxy.ProxyRequest("comment", "/api/v1/comments/:movieId"))
			write.DELETE("/:id", proxy.ProxyRequest("comment", "/api/v1/comments/:id"))
		}
	}
}
