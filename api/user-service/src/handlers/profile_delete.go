package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

func DeleteProfileHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid user ID")
		return
	}

	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}
	if uint(id) != userID {
		utils.RespondError(c, http.StatusForbidden, "cannot delete another user's profile")
		return
	}

	var user models.User
	if err := conf.DB.First(&user, id).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}

	if err := conf.DB.Delete(&user).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to delete profile")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "Profile deleted successfully"})
}
