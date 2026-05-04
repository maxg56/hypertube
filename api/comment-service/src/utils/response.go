package utils

import "github.com/gin-gonic/gin"

type StandardResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func RespondSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, StandardResponse{Success: true, Data: data})
}

func RespondError(c *gin.Context, status int, message string) {
	c.JSON(status, StandardResponse{Success: false, Error: message})
}
