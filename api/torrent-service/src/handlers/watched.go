package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"torrent-service/src/services"
	"torrent-service/src/utils"
)

// WatchedHandler handles GET /api/v1/movies/:id/watched
func WatchedHandler(c *gin.Context) {
	userID, ok := userIDFromHeader(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "missing or invalid X-User-ID")
		return
	}

	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil || movieID <= 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	watched, err := services.IsWatched(userID, movieID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "database error")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"watched": watched})
}

func userIDFromHeader(c *gin.Context) (int, bool) {
	raw := c.GetHeader("X-User-ID")
	if raw == "" {
		return 0, false
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
