package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

type Pagination struct {
	Total  int64 `json:"total"`
	Limit  int   `json:"limit"`
	Offset int   `json:"offset"`
}

type PaginationParams struct {
	Limit  int
	Offset int
}

func RespondSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, StandardResponse{Success: true, Data: data})
}

func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, StandardResponse{Success: false, Error: message})
}

func RespondPaginated(c *gin.Context, status int, data interface{}, total int64, p PaginationParams) {
	c.JSON(status, PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: Pagination{
			Total:  total,
			Limit:  p.Limit,
			Offset: p.Offset,
		},
	})
}

func ParsePagination(c *gin.Context) PaginationParams {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}
	return PaginationParams{Limit: limit, Offset: offset}
}
