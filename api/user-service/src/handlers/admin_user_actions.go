package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

// AdminPromoteUserHandler handles PUT /api/v1/admin/users/:id/role
func AdminPromoteUserHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var body struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || (body.Role != "admin" && body.Role != "user") {
		utils.RespondError(c, http.StatusBadRequest, "role must be 'admin' or 'user'")
		return
	}

	result := conf.DB.Model(&models.User{}).Where("id = ?", id).Update("role", body.Role)
	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to update role")
		return
	}
	if result.RowsAffected == 0 {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "role updated"})
}

// AdminDeleteUserHandler handles DELETE /api/v1/admin/users/:id
func AdminDeleteUserHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	callerID := c.GetHeader("X-User-ID")
	if callerID == strconv.FormatUint(id, 10) {
		utils.RespondError(c, http.StatusForbidden, "cannot delete your own account")
		return
	}

	var user models.User
	if err := conf.DB.First(&user, id).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}

	if user.AvatarURL != "" && strings.HasPrefix(user.AvatarURL, "/api/v1/users/avatars/") {
		avatarDir := os.Getenv("AVATAR_DIR")
		if avatarDir == "" {
			avatarDir = "/data/avatars"
		}
		filename := filepath.Base(user.AvatarURL)
		os.Remove(filepath.Join(avatarDir, filename)) //nolint:errcheck
	}

	if err := conf.DB.Delete(&user).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to delete user")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "user deleted"})
}

// AdminUpdateUsernameHandler handles PUT /api/v1/admin/users/:id/username
func AdminUpdateUsernameHandler(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		utils.RespondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var body struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Username) == "" {
		utils.RespondError(c, http.StatusBadRequest, "username is required")
		return
	}

	result := conf.DB.Model(&models.User{}).Where("id = ?", id).Update("username", strings.TrimSpace(body.Username))
	if result.Error != nil {
		utils.RespondError(c, http.StatusConflict, "username already taken or invalid")
		return
	}
	if result.RowsAffected == 0 {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"message": "username updated"})
}
