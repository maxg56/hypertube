package handlers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"comment-service/src/conf"
	"comment-service/src/models"
	"comment-service/src/utils"
)

var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

func sanitizeText(s string) string {
	return strings.TrimSpace(htmlTagRegex.ReplaceAllString(s, ""))
}

func requireUserID(c *gin.Context) (int, bool) {
	s := c.GetHeader("X-User-ID")
	if s == "" {
		utils.RespondError(c, http.StatusUnauthorized, "missing user id")
		return 0, false
	}
	id, err := strconv.Atoi(s)
	if err != nil {
		utils.RespondError(c, http.StatusUnauthorized, "invalid user id")
		return 0, false
	}
	return id, true
}

// commentWithMeta is used by joins that need user + movie data alongside a comment.
type commentWithMeta struct {
	ID         uint      `gorm:"column:id"`
	UserID     int       `gorm:"column:user_id"`
	MovieID    int       `gorm:"column:movie_id"`
	Content    string    `gorm:"column:content"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
	Username   string    `gorm:"column:username"`
	AvatarURL  string    `gorm:"column:avatar_url"`
	TMDbID     int       `gorm:"column:tmdb_id"`
	MovieTitle string    `gorm:"column:movie_title"`
}

func ListCommentsByUser(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userId"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid user id")
		return
	}

	p := utils.ParsePagination(c)

	var total int64
	conf.DB.Model(&models.Comment{}).Where("user_id = ?", userID).Count(&total)

	var rows []commentWithMeta
	conf.DB.Model(&models.Comment{}).
		Select("comments.id, comments.user_id, comments.movie_id, comments.content, comments.created_at, "+
			"users.username, users.avatar_url, movies.tmdb_id, movies.title AS movie_title").
		Joins("JOIN users ON users.id = comments.user_id").
		Joins("JOIN movies ON movies.id = comments.movie_id").
		Where("comments.user_id = ?", userID).
		Order("comments.created_at DESC").
		Limit(p.Limit).Offset(p.Offset).
		Find(&rows)

	type UserComment struct {
		ID         uint   `json:"id"`
		Content    string `json:"content"`
		CreatedAt  string `json:"created_at"`
		MovieID    int    `json:"movie_id"`
		TMDbID     int    `json:"tmdb_id"`
		MovieTitle string `json:"title"`
	}

	result := make([]UserComment, 0, len(rows))
	for _, r := range rows {
		result = append(result, UserComment{
			ID:         r.ID,
			Content:    r.Content,
			CreatedAt:  r.CreatedAt.Format(time.RFC3339),
			MovieID:    r.MovieID,
			TMDbID:     r.TMDbID,
			MovieTitle: r.MovieTitle,
		})
	}

	utils.RespondPaginated(c, http.StatusOK, result, total, p)
}

func ListAllComments(c *gin.Context) {
	p := utils.ParsePagination(c)

	var total int64
	conf.DB.Model(&models.Comment{}).Count(&total)

	var rows []commentWithMeta
	conf.DB.Model(&models.Comment{}).
		Select("comments.id, comments.user_id, comments.content, comments.created_at, "+
			"users.username, users.avatar_url, movies.tmdb_id, movies.title AS movie_title").
		Joins("JOIN users ON users.id = comments.user_id").
		Joins("JOIN movies ON movies.id = comments.movie_id").
		Order("comments.created_at DESC").
		Limit(p.Limit).Offset(p.Offset).
		Find(&rows)

	type AllComment struct {
		ID         uint   `json:"id"`
		UserID     int    `json:"user_id"`
		Username   string `json:"username"`
		AvatarURL  string `json:"avatar_url,omitempty"`
		Content    string `json:"content"`
		CreatedAt  string `json:"created_at"`
		TMDbID     int    `json:"tmdb_id"`
		MovieTitle string `json:"title"`
	}

	result := make([]AllComment, 0, len(rows))
	for _, r := range rows {
		result = append(result, AllComment{
			ID:         r.ID,
			UserID:     r.UserID,
			Username:   r.Username,
			AvatarURL:  r.AvatarURL,
			Content:    r.Content,
			CreatedAt:  r.CreatedAt.Format(time.RFC3339),
			TMDbID:     r.TMDbID,
			MovieTitle: r.MovieTitle,
		})
	}

	utils.RespondPaginated(c, http.StatusOK, result, total, p)
}

func GetComment(c *gin.Context) {
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid comment id")
		return
	}

	var row commentWithMeta
	res := conf.DB.Model(&models.Comment{}).
		Select("comments.id, comments.user_id, comments.content, comments.created_at, comments.updated_at, "+
			"users.username, users.avatar_url, movies.tmdb_id, movies.title AS movie_title").
		Joins("JOIN users ON users.id = comments.user_id").
		Joins("JOIN movies ON movies.id = comments.movie_id").
		Where("comments.id = ?", commentID).
		First(&row)

	if res.Error != nil {
		utils.RespondError(c, http.StatusNotFound, "comment not found")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"id":         row.ID,
		"user_id":    row.UserID,
		"username":   row.Username,
		"avatar_url": row.AvatarURL,
		"content":    row.Content,
		"created_at": row.CreatedAt.Format(time.RFC3339),
		"updated_at": row.UpdatedAt.Format(time.RFC3339),
		"tmdb_id":    row.TMDbID,
		"title":      row.MovieTitle,
	})
}

func ListComments(c *gin.Context) {
	tmdbID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	p := utils.ParsePagination(c)

	var movie models.Movie
	if err := conf.DB.Where("tmdb_id = ?", tmdbID).First(&movie).Error; err != nil {
		utils.RespondPaginated(c, http.StatusOK, []models.CommentResponse{}, 0, p)
		return
	}

	var total int64
	conf.DB.Model(&models.Comment{}).Where("movie_id = ?", movie.ID).Count(&total)

	var rows []commentWithMeta
	conf.DB.Model(&models.Comment{}).
		Select("comments.id, comments.user_id, comments.content, comments.created_at, "+
			"users.username, users.avatar_url").
		Joins("JOIN users ON users.id = comments.user_id").
		Where("comments.movie_id = ?", movie.ID).
		Order("comments.created_at DESC").
		Limit(p.Limit).Offset(p.Offset).
		Find(&rows)

	result := make([]models.CommentResponse, 0, len(rows))
	for _, r := range rows {
		result = append(result, models.CommentResponse{
			ID:        r.ID,
			UserID:    r.UserID,
			Username:  r.Username,
			AvatarURL: r.AvatarURL,
			Content:   r.Content,
			CreatedAt: r.CreatedAt,
		})
	}

	utils.RespondPaginated(c, http.StatusOK, result, total, p)
}

type createCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
	Title   string `json:"title"`
}

func CreateComment(c *gin.Context) {
	tmdbID, err := strconv.Atoi(c.Param("movieId"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid movie id")
		return
	}

	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var req createCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "content is required")
		return
	}

	req.Content = sanitizeText(req.Content)
	if req.Content == "" {
		utils.RespondError(c, http.StatusBadRequest, "comment content cannot be empty")
		return
	}

	title := req.Title
	if title == "" {
		title = "Unknown"
	}

	movie := models.Movie{TMDbID: tmdbID, Title: title}
	if err := conf.DB.Where(models.Movie{TMDbID: tmdbID}).FirstOrCreate(&movie).Error; err != nil {
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

type updateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1,max=2000"`
}

func UpdateComment(c *gin.Context) {
	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "invalid comment id")
		return
	}

	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var req updateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "content is required")
		return
	}

	req.Content = sanitizeText(req.Content)
	if req.Content == "" {
		utils.RespondError(c, http.StatusBadRequest, "comment content cannot be empty")
		return
	}

	var comment models.Comment
	if err := conf.DB.First(&comment, commentID).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "comment not found")
		return
	}

	if comment.UserID != userID {
		utils.RespondError(c, http.StatusForbidden, "cannot edit another user's comment")
		return
	}

	if err := conf.DB.Model(&comment).Update("content", req.Content).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to update comment")
		return
	}

	var user models.User
	conf.DB.First(&user, userID)

	utils.RespondSuccess(c, http.StatusOK, models.CommentResponse{
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

	userID, ok := requireUserID(c)
	if !ok {
		return
	}

	var comment models.Comment
	if err := conf.DB.First(&comment, commentID).Error; err != nil {
		utils.RespondError(c, http.StatusNotFound, "comment not found")
		return
	}

	role := c.GetHeader("X-User-Role")
	if comment.UserID != userID && role != "admin" {
		utils.RespondError(c, http.StatusForbidden, "cannot delete another user's comment")
		return
	}

	if err := conf.DB.Delete(&comment).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to delete comment")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{"deleted": true})
}
