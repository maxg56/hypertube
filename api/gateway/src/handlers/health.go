package handlers

import (
	"net/http"
	"time"

	"gateway/src/services"
	"github.com/gin-gonic/gin"
)

type HealthCheckResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Services  map[string]string `json:"services"`
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthCheckResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
		Services:  services.GetServicesStatus(),
	})
}
