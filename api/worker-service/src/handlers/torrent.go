package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"worker-service/src/services"
	"worker-service/src/utils"
)

type deleteTorrentRequest struct {
	TorrentID int    `json:"torrent_id" binding:"required"`
	InfoHash  string `json:"info_hash"  binding:"required"`
	FilePath  string `json:"file_path"`
}

func DeleteTorrentHandler(c *gin.Context) {
	var req deleteTorrentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	utils.DeleteFromDisk(
		utils.DownloadDir()+"/"+req.InfoHash,
		req.FilePath,
	)

	if err := services.MarkTorrentAsDeleted(req.TorrentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db update failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}
