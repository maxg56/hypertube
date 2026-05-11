package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/models"
	"user-service/src/utils"
)

type adminUserRow struct {
	ID              uint             `json:"id"`
	Username        string           `json:"username"`
	Email           string           `json:"email"`
	FirstName       string           `json:"first_name"`
	LastName        string           `json:"last_name"`
	AvatarURL       string           `json:"avatar_url"`
	Role            models.UserRole  `json:"role"`
	EmailVerified   bool             `json:"email_verified"`
	CreatedAt       time.Time        `json:"created_at"`
	FilmsWatched    int64            `json:"films_watched"`
	FilmsDownloaded int64            `json:"films_downloaded"`
}

// AdminListUsersHandler handles GET /api/v1/admin/users
func AdminListUsersHandler(c *gin.Context) {
	pagination := utils.ParsePaginationParams(c)

	var total int64
	if err := conf.DB.Table("users").Count(&total).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to count users")
		return
	}

	var rows []adminUserRow
	err := conf.DB.Raw(`
		SELECT
			u.id,
			u.username,
			u.email,
			u.first_name,
			u.last_name,
			COALESCE(u.avatar_url, '') AS avatar_url,
			u.role,
			u.email_verified,
			u.created_at,
			COUNT(DISTINCT wh.movie_id) AS films_watched,
			COUNT(DISTINCT CASE WHEN t.status = 'ready' THEN wh.movie_id END) AS films_downloaded
		FROM users u
		LEFT JOIN watch_history wh ON wh.user_id = u.id
		LEFT JOIN torrents t ON t.movie_id = wh.movie_id
		GROUP BY u.id
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`, pagination.Limit, pagination.Offset).Scan(&rows).Error
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"users":      rows,
		"pagination": utils.NewPagination(total, pagination.Limit, pagination.Offset),
	})
}
