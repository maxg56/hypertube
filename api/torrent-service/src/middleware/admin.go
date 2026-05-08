package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"torrent-service/src/utils"
)

// AdminMiddleware rejects requests whose X-User-Role header is not "admin".
// Must be placed after any auth middleware that sets X-User-ID.
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("X-User-Role") != "admin" {
			utils.RespondError(c, http.StatusForbidden, "forbidden: admin access required")
			c.Abort()
			return
		}
		c.Next()
	}
}
