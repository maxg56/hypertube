package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

func GetOwnProfileHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	var user models.User
	if err := conf.DB.First(&user, userID).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"profile": user})
}
