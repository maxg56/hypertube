package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminMiddleware rejects requests from non-admin users.
// Must be chained after JWTMiddleware.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get(CtxUserRoleKey)
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden: admin access required"})
			return
		}
		c.Next()
	}
}
