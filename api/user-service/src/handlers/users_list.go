package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"user-service/src/conf"
	"user-service/src/utils"
)

type userSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func ListUsersHandler(c *gin.Context) {
	pagination := utils.ParsePaginationParams(c)

	var total int64
	if err := conf.DB.Model(&userSummary{}).Table("users").Count(&total).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to count users")
		return
	}

	var users []userSummary
	if err := conf.DB.Table("users").
		Select("id, username").
		Limit(pagination.Limit).
		Offset(pagination.Offset).
		Find(&users).Error; err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	utils.RespondSuccess(c, http.StatusOK, gin.H{
		"users":      users,
		"pagination": utils.NewPagination(total, pagination.Limit, pagination.Offset),
	})
}
