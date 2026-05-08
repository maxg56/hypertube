package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"torrent-service/src/services"
	"torrent-service/src/utils"
)

// AdminListFilmsHandler handles GET /api/v1/admin/films
func AdminListFilmsHandler(c *gin.Context) {
	limit, offset := parsePagination(c)

	films, total, err := services.ListAdminFilms(limit, offset)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to fetch films: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"films": films,
		"pagination": gin.H{
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
	})
}

// AdminDeleteFilmHandler handles DELETE /api/v1/admin/films/:id
func AdminDeleteFilmHandler(c *gin.Context) {
	id, err := parseTorrentID(c)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid id")
		return
	}

	if err := services.DeleteAdminFilm(id); err != nil {
		if err.Error() == "not found" {
			utils.RespondError(c, http.StatusNotFound, "film not found")
			return
		}
		utils.RespondError(c, http.StatusInternalServerError, "failed to delete film: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "film deleted successfully"})
}

// AdminReDownloadFilmHandler handles POST /api/v1/admin/films/:id/download
func AdminReDownloadFilmHandler(c *gin.Context) {
	id, err := parseTorrentID(c)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid id")
		return
	}

	infoHash, err := services.ReDownloadFilm(id)
	if err != nil {
		if err.Error() == "not found" {
			utils.RespondError(c, http.StatusNotFound, "film not found")
			return
		}
		utils.RespondError(c, http.StatusBadRequest, "failed to re-trigger download: "+err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusAccepted, gin.H{
		"info_hash": infoHash,
		"status":    "downloading",
		"message":   "torrent re-download triggered",
	})
}

func parseTorrentID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func parsePagination(c *gin.Context) (int, int) {
	limit := 20
	offset := 0
	if l, err := strconv.Atoi(c.DefaultQuery("limit", "20")); err == nil && l > 0 {
		limit = l
	}
	if o, err := strconv.Atoi(c.DefaultQuery("offset", "0")); err == nil && o >= 0 {
		offset = o
	}
	return limit, offset
}
