package handlers

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"torrent-service/src/services"
	"torrent-service/src/utils"
)

// SubtitleHandler handles GET /api/v1/movies/:id/subtitles/:lang
// Returns a WebVTT subtitle file for the given movie and language code.
func SubtitleHandler(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil || movieID <= 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	lang := c.Param("lang")
	if lang == "" {
		utils.RespondError(c, http.StatusBadRequest, "missing language")
		return
	}

	path, err := services.FetchSubtitle(movieID, lang)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, "subtitles unavailable: "+err.Error())
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to read subtitle file")
		return
	}

	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "text/vtt; charset=utf-8", data)
}
