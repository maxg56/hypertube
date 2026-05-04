package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"comment-service/src/conf"
	"comment-service/src/models"
	"comment-service/src/utils"
)

type createCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
	Title   string `json:"title"`
}

func ListComments(c *gin.Context) {
	tmdbID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	var movie models.Movie
	if err := conf.DB.Where("tmdb_id = ?", tmdbID).First(&movie).Error; err != nil {
		utils.RespondSuccess(c, http.StatusOK, []models.CommentResponse{})
		return
	}

	type row struct {
		ID        uint   `gorm:"column:id"`
		UserID    int    `gorm:"column:user_id"`
		Username  string `gorm:"column:username"`
		AvatarURL string `gorm:"column:avatar_url"`
		Content   string `gorm:"column:content"`
		CreatedAt string `gorm:"column:created_at"`
	}

	var rows []row
	conf.DB.Raw(`
		SELECT c.id, c.user_id, u.username, u.avatar_url, c.content, c.created_at
		FROM comments c
		INNER JOIN users u ON u.id = c.user_id
		WHERE c.movie_id = ?
		ORDER BY c.created_at DESC
	`, movie.ID).Scan(&rows)

	result := make([]models.CommentResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, models.CommentResponse{
			ID:        r.ID,
			UserID:    r.UserID,
			Username:  r.Username,
			AvatarURL: r.AvatarURL,
			Content:   r.Content,
		})
	}

	utils.RespondSuccess(c, http.StatusOK, result)
}

func CreateComment(c *gin.Context) {
	tmdbID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.RespondError(c, http.StatusUnauthorized, "missing user id")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "invalid user id")
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "content is required")
		return
	}

	var movie models.Movie
	title := req.Title
	if title == "" {
		title = "Unknown"
	}

	result := conf.DB.Raw(`
		INSERT INTO movies (tmdb_id, title)
		VALUES (?, ?)
		ON CONFLICT (tmdb_id) DO UPDATE SET cached_at = CURRENT_TIMESTAMP
		RETURNING id, tmdb_id, title
	`, tmdbID, title).Scan(&movie)
	if result.Error != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to resolve movie")
		return
	}

	comment := models.Comment{
		MovieID: int(movie.ID),
		UserID:  userID,
		Content: req.Content,
	}
	if err := conf.DB.Create(&comment).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to create comment")
		return
	}

	var user models.User
	conf.DB.First(&user, userID)

	utils.RespondSuccess(c, http.StatusCreated, models.CommentResponse{
		ID:        comment.ID,
		UserID:    userID,
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
	})
}

func DeleteComment(c *gin.Context) {
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid comment id")
		return
	}

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		utils.RespondError(c, http.StatusUnauthorized, "missing user id")
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "invalid user id")
		return
	}

	var comment models.Comment
	if err := conf.DB.First(&comment, commentID).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "comment not found")
		return
	}

	if comment.UserID != userID {
		utils.RespondError(c, http.StatusForbidden, "cannot delete another user's comment")
		return
	}

	if err := conf.DB.Delete(&comment).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to delete comment")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"deleted": true})
}
