package handlers

import (
	"fmt"
	"io"
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

const maxAvatarSize = 5 << 20 // 5 MB

var allowedMIMETypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func UploadAvatarHandler(c *gin.Context) {
	userID, err := utils.GetAuthenticatedUserID(c)
	if err != nil {
		return
	}

	file, _, err := c.Request.FormFile("avatar")
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "avatar file is required")
		return
	}
	defer file.Close()

	// Read up to maxAvatarSize+1 to detect oversized files
	limited := io.LimitReader(file, maxAvatarSize+1)
	buf, err := io.ReadAll(limited)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to read file")
		return
	}
	if len(buf) > maxAvatarSize {
		utils.RespondError(c, http.StatusBadRequest, "avatar file exceeds 5 MB limit")
		return
	}

	// Validate MIME type from actual file content (magic bytes)
	mime := http.DetectContentType(buf)
	// DetectContentType may return "image/jpeg" or "image/jpeg; charset=..." style
	mime = strings.SplitN(mime, ";", 2)[0]
	ext, ok := allowedMIMETypes[mime]
	if !ok {
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

	if err := os.WriteFile(dst, buf, 0644); err != nil {
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
