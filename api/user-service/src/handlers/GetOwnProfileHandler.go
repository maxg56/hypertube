package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

func GetOwnProfileHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	if userID == 0 {
		utils.RespondError(c, http.StatusUnauthorized, "user not authenticated")
		return
	}

	var user models.User
	if err := conf.DB.First(&user, uint(userID)).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"profile": user})
}
