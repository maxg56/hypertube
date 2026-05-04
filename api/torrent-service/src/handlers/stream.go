package handlers

import (
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"

	"torrent-service/src/models"
	"torrent-service/src/services"
	"torrent-service/src/utils"
)

func StreamHandler(c *gin.Context) {
	hash := c.Param("id")
	if hash == "" {
		utils.RespondError(c, http.StatusBadRequest, "missing info hash")
		return
	}

	record, err := services.GetRecord(hash)
	if err != nil {
		utils.RespondError(c, http.StatusNotFound, "torrent not found")
		return
	}

	switch record.Status {
	case models.StatusPending:
		c.Header("Retry-After", "5")
		utils.RespondError(c, http.StatusAccepted, "torrent is pending, retry shortly")
		return
	case models.StatusError:
		utils.RespondError(c, http.StatusServiceUnavailable, "torrent failed: "+record.ErrorMsg)
		return
	}

	result, err := services.GetTorrentReader(hash)
	if err != nil {
		utils.RespondError(c, http.StatusServiceUnavailable, "cannot open torrent: "+err.Error())
		return
	}

	contentType := mime.TypeByExtension(filepath.Ext(result.FileName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	if result.Size > 0 {
		c.Header("X-Content-Length", fmt.Sprint(result.Size))
	}

	http.ServeContent(c.Writer, c.Request, result.FileName, time.Time{}, result.Reader)
}
