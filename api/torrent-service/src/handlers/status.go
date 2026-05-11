package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"torrent-service/src/services"
	"torrent-service/src/types"
	"torrent-service/src/utils"
)

func StatusHandler(c *gin.Context) {
	hash := strings.ToLower(c.Param("id"))
	if hash == "" {
		utils.RespondError(c, http.StatusBadRequest, "missing info hash")
		return
	}

	record, err := services.GetRecord(hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.RespondError(c, http.StatusNotFound, "torrent not found")
		} else {
			utils.RespondError(c, http.StatusInternalServerError, err.Error())
		}
		return
	}

	utils.RespondSuccess(c, http.StatusOK, types.StatusResponse{
		InfoHash:   record.InfoHash,
		Status:     string(record.Status),
		Progress:   record.Progress,
		Downloaded: record.Downloaded,
		FileSize:   record.FileSize,
		FilePath:   record.FilePath,
		ErrorMsg:   record.ErrorMsg,
	})
}
