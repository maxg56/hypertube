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
		comment.GET("", proxy.ProxyRequest("comment", "/api/v1/comments"))
		comment.GET("/user/:userId", proxy.ProxyRequest("comment", "/api/v1/comments/user/:userId"))
		comment.GET("/:id", proxy.ProxyRequest("comment", "/api/v1/comments/:id"))
		comment.PATCH("/:id", proxy.ProxyRequest("comment", "/api/v1/comments/:id"))
		comment.DELETE("/:id", proxy.ProxyRequest("comment", "/api/v1/comments/:id"))
	}

	movies := r.Group("/api/v1/movies")
	movies.Use(middleware.JWTMiddleware())
	{
		movies.GET("/:id/comments", proxy.ProxyRequest("comment", "/api/v1/movies/:id/comments"))
		movies.POST("/:id/comments", proxy.ProxyRequest("comment", "/api/v1/movies/:id/comments"))
	}
}
