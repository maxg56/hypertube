package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"user-service/src/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from header (set by gateway)
		userIDStr := c.GetHeader("X-User-ID")
		
		if userIDStr == "" {
			utils.RespondError(c, http.StatusUnauthorized, "user not authenticated")
			c.Abort()
			return
		}

		// Validate user ID
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil || userID == 0 {
			utils.RespondError(c, http.StatusUnauthorized, "user not authenticated")
			c.Abort()
			return
		}

		// Store user ID in context as uint to match the DB model type
		c.Set("user_id", uint(userID))
		c.Next()
	}
}