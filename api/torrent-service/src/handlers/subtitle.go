package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"torrent-service/src/services"
	"torrent-service/src/utils"
)

// SubtitleAvailableHandler handles GET /api/v1/movies/:id/subtitles
// Returns the list of language codes that are either cached on disk or available
// for download from OpenSubtitles for this movie.
func SubtitleAvailableHandler(c *gin.Context) {
	movieID, err := strconv.Atoi(c.Param("id"))
	if err != nil || movieID <= 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	all := map[string]bool{}

	dir := filepath.Join(services.SubtitleCacheDir(), fmt.Sprintf("%d", movieID))
	if entries, err := os.ReadDir(dir); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".vtt") {
				all[strings.TrimSuffix(e.Name(), ".vtt")] = true
			}
		}
	}

	for _, lang := range services.ListAvailableLanguages(movieID) {
		all[lang] = true
	}

	langs := make([]string, 0, len(all))
	for l := range all {
		langs = append(langs, l)
	}
	utils.RespondSuccess(c, http.StatusOK, gin.H{"languages": langs})
}

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
