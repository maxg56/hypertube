package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"torrent-service/src/services"
	"torrent-service/src/types"
	"torrent-service/src/utils"
)

func DownloadHandler(c *gin.Context) {
	var req types.DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	infoHash, err := services.StartDownload(req.MagnetURI, req.MovieID)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}

	utils.RespondSuccess(c, http.StatusAccepted, types.DownloadResponse{
		InfoHash: infoHash,
		Status:   "downloading",
		Message:  "torrent download started",
	})
}
