package handlers

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

func UpdateProfileHandler(c *gin.Context) {
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
		utils.RespondError(c, http.StatusForbidden, "cannot update another user's profile")
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid payload: "+err.Error())
		return
	}

	var user models.User
	if err := conf.DB.First(&user, id).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.AvatarURL != nil {
		parsed, err := url.ParseRequestURI(*req.AvatarURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			utils.RespondError(c, http.StatusBadRequest, "invalid avatar URL")
			return
		}
		user.AvatarURL = *req.AvatarURL
	}
	if req.Language != nil {
		allowed := map[string]bool{"fr": true, "en": true}
		if !allowed[*req.Language] {
			utils.RespondError(c, http.StatusBadRequest, "unsupported language")
			return
		}
		user.Language = *req.Language
	}

	if err := conf.DB.Save(&user).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to update profile")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"profile": user,
	})
}
