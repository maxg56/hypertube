package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

type userSearchResult struct {
	ID        uint   `json:"id"         gorm:"column:id"`
	Username  string `json:"username"   gorm:"column:username"`
	AvatarURL string `json:"avatar_url" gorm:"column:avatar_url"`
	FirstName string `json:"first_name" gorm:"column:first_name"`
	LastName  string `json:"last_name"  gorm:"column:last_name"`
}

// SearchUsersHandler handles GET /api/v1/users/search?q=<username> (public)
func SearchUsersHandler(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		utils.RespondSuccess(c, http.StatusOK, gin.H{"users": []struct{}{}})
		return
	}

	var results []userSearchResult
	if err := conf.DB.Model(&models.User{}).
		Select("id, username, avatar_url, first_name, last_name").
		Where("is_public = true AND username ILIKE ?", "%"+q+"%").
		Limit(20).
		Find(&results).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to search users")
		return
	}

	if results == nil {
		results = []userSearchResult{}
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"users": results})
}
