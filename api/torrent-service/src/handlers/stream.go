package handlers

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"torrent-service/src/models"
	"torrent-service/src/services"
	"torrent-service/src/utils"
)

func init() {
	mime.AddExtensionType(".mkv", "video/x-matroska")
	mime.AddExtensionType(".webm", "video/webm")
	mime.AddExtensionType(".mp4", "video/mp4")
	mime.AddExtensionType(".avi", "video/x-msvideo")
	mime.AddExtensionType(".mov", "video/quicktime")
	mime.AddExtensionType(".ogg", "video/ogg")
	mime.AddExtensionType(".m4v", "video/mp4")
}

func StreamHandler(c *gin.Context) {
	hash := strings.ToLower(c.Param("id"))
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
		log.Printf("torrent %s failed: %s", hash, record.ErrorMsg)
		utils.RespondError(c, http.StatusServiceUnavailable, record.ErrorMsg)
		return
	}

	userID, _ := userIDFromHeader(c)
	if userID > 0 && record.MovieID > 0 {
		services.RecordWatch(userID, record.MovieID) //nolint:errcheck
	}

	result, err := services.GetTorrentReader(hash)
	if err != nil {
		log.Printf("cannot open torrent %s: %v", hash, err)
		utils.RespondError(c, http.StatusServiceUnavailable, "torrent not available")
		return
	}

	// Bug 1 fix: single NeedsTranscoding check (duplicate block removed).
	if services.NeedsTranscoding(result.FileName) {
		serveTranscoded(c, result, userID, hash)
		return
	}

	c.Header("Accept-Ranges", "bytes")
	if result.Size > 0 {
		c.Header("X-Content-Length", fmt.Sprint(result.Size))
	}
	http.ServeContent(c.Writer, c.Request, result.FileName, time.Time{}, result.Reader)
}

func serveTranscoded(c *gin.Context, result services.ReaderResult, userID int, infoHash string) {
	var codecInfo *services.CodecInfo
	if result.FilePath != "" {
		codecInfo, _ = services.ProbeCodecs(result.FilePath)
	}

	job, err := services.StartTranscode(result.Reader, codecInfo)
	if err != nil {
		// Bug 2 fix: single RespondError call (log first, then respond once).
		log.Printf("transcoding error for %s: %v", result.FileName, err)
		utils.RespondError(c, http.StatusInternalServerError, "transcoding error: "+err.Error())
		return
	}
	// Bug 3 fix: session ties ffmpeg lifecycle to the request context.
	// Release runs first (LIFO), killing ffmpeg, then Wait collects exit status.
	defer services.Sessions.Release(userID, infoHash)
	defer job.Cmd.Wait() //nolint:errcheck
	services.Sessions.Acquire(userID, infoHash, job, c.Request.Context())

	c.Header("Content-Type", job.ContentType)
	c.Header("Cache-Control", "no-cache")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)
	io.Copy(c.Writer, job.Reader) //nolint:errcheck
}
