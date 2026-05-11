package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"torrent-service/src/services"
	"torrent-service/src/utils"
)

// GetProgressHandler handles GET /api/v1/movies/:id/progress
// :id is the TMDB movie ID.
func GetProgressHandler(c *gin.Context) {
	userID, ok := userIDFromHeader(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "missing or invalid X-User-ID")
		return
	}

	tmdbID, err := strconv.Atoi(c.Param("id"))
	if err != nil || tmdbID <= 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	localID, err := services.ResolveLocalMovieID(tmdbID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "could not resolve movie")
		return
	}

	sec, err := services.GetProgress(userID, localID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "database error")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"progress_sec": sec})
}

// SaveProgressHandler handles PUT /api/v1/movies/:id/progress
// :id is the TMDB movie ID.
func SaveProgressHandler(c *gin.Context) {
	userID, ok := userIDFromHeader(c)
	if !ok {
		utils.RespondError(c, http.StatusUnauthorized, "missing or invalid X-User-ID")
		return
	}

	tmdbID, err := strconv.Atoi(c.Param("id"))
	if err != nil || tmdbID <= 0 {
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

	localID, err := services.ResolveLocalMovieID(tmdbID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "could not resolve movie")
		return
	}

	if err := services.SaveProgress(userID, localID, body.ProgressSec); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "database error")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"progress_sec": body.ProgressSec})
}
