package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"user-service/src/services"
	"user-service/src/utils"
)

func GetUserOnlineStatusHandler(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	presenceService := services.NewPresenceService()
	presence, err := presenceService.GetUserPresence(uint(userID))
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to get user status")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"id":            presence.UserID,
		"is_online":     presence.IsOnline,
		"last_seen":     presence.LastSeen,
		"last_activity": presence.LastActivity,
	})
}
