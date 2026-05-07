package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

func UploadAvatarHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowed[ext] {
		utils.RespondError(c, http.StatusBadRequest, "unsupported image format")
		return
	}

	avatarDir := os.Getenv("AVATAR_DIR")
	if avatarDir == "" {
		avatarDir = "/data/avatars"
	}
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to create upload directory")
		return
	}

	filename := fmt.Sprintf("%d_%d%s", userID, time.Now().UnixNano(), ext)
	dst := filepath.Join(avatarDir, filename)

	buf := make([]byte, 5<<20)
	n, _ := file.Read(buf)
	if err := os.WriteFile(dst, buf[:n], 0644); err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to save avatar")
		return
	}

	avatarURL := fmt.Sprintf("/api/v1/users/avatars/%s", filename)

	var user models.User
	if err := conf.DB.First(&user, userID).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "user not found")
		return
	}

	if user.AvatarURL != "" && strings.HasPrefix(user.AvatarURL, "/api/v1/users/avatars/") {
		old := filepath.Join(avatarDir, filepath.Base(user.AvatarURL))
		os.Remove(old)
	}

	user.AvatarURL = avatarURL
	if err := conf.DB.Save(&user).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to update avatar")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"avatar_url": avatarURL})
}
