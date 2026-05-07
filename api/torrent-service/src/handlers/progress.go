package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"torrent-service/src/services"
	"torrent-service/src/utils"
)

// GetProgressHandler handles GET /api/v1/movies/:id/progress
func GetProgressHandler(c *gin.Context) {
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

	sec, err := services.GetProgress(userID, movieID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "database error")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"progress_sec": sec})
}

// SaveProgressHandler handles PUT /api/v1/movies/:id/progress
func SaveProgressHandler(c *gin.Context) {
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

	var body struct {
		ProgressSec int `json:"progress_sec"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.ProgressSec < 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid progress_sec")
		return
	}

	if err := services.SaveProgress(userID, movieID, body.ProgressSec); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "database error")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"progress_sec": body.ProgressSec})
}
