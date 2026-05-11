package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"torrent-service/src/services"
	"torrent-service/src/utils"
)

func AdminStatsHandler(c *gin.Context) {
	stats, err := services.GetAdminStats()
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to fetch stats: "+err.Error())
		return
	}
	utils.RespondSuccess(c, http.StatusOK, stats)
}
