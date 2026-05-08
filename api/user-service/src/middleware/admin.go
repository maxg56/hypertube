package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"user-service/src/utils"
)

// AdminMiddleware rejects requests from non-admin users.
// Must be chained after AuthMiddleware.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetHeader("X-User-Role")
		if role != "admin" {
			utils.RespondError(c, http.StatusForbidden, "forbidden: admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
