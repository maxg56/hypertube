package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAuthenticatedUserID retrieves and validates the authenticated user ID from the Gin context
func GetAuthenticatedUserID(c *gin.Context) (uint, error) {
	authenticatedUserID, exists := c.Get("user_id")
	if !exists {
		RespondError(c, http.StatusUnauthorized, "user not authenticated")
		return 0, ErrUnauthorized
	}

	userID, ok := authenticatedUserID.(int)
	if !ok {
		RespondError(c, http.StatusUnauthorized, "user not authenticated")
		return 0, ErrUnauthorized
	}
	return uint(userID), nil
}

// Custom errors
var (
	ErrUnauthorized = NewAppError("user not authenticated", http.StatusUnauthorized)
)