package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"gateway/src/config"
	"gateway/src/handlers"
	"gateway/src/middleware"
	"gateway/src/routes"
	"gateway/src/services"
	"gateway/src/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg, err := config.LoadAndValidateConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	services.InitServices()

	if cfg.RateLimitEnabled {
		middleware.InitRateLimiter(cfg.RateLimitRPS)
		log.Printf("Rate limiter initialized: %d RPS per client", cfg.RateLimitRPS)
	}

	if err := utils.InitRedis(); err != nil {
		log.Printf("Failed to initialize Redis: %v", err)
		log.Println("Redis initialization failed - JWT blacklisting will be disabled")
	} else {
		log.Println("Redis initialized successfully for JWT blacklisting")
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(handlers.CORSMiddleware())
	r.Use(middleware.RateLimitMiddleware())

	r.GET("/health", handlers.HealthCheck)
	r.GET("/api/health", handlers.HealthCheck)

	routes.SetupAuthRoutes(r)
	routes.SetupUserRoutes(r)
	routes.SetupLibraryRoutes(r)
	routes.SetupTorrentRoutes(r)
	routes.SetupCommentRoutes(r)

	log.Printf("Gateway starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
