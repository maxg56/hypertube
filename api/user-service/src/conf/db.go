package conf

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"user-service/src/models"
)

var DB *gorm.DB

func InitDB() {
	host := getEnvWithDefault("DB_HOST", "localhost")
	port := getEnvWithDefault("DB_PORT", "5432")
	user := getEnvWithDefault("DB_USER", "postgres")
	password := getEnvWithDefault("DB_PASSWORD", "password")
	dbname := getEnvWithDefault("DB_NAME", "hypertube")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	const maxRetries = 10
	var err error
	for i := range maxRetries {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err == nil {
			DB = DB.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Warn)})
			break
		}
		log.Printf("Database not ready, retrying in 2s (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if getEnvWithDefault("AUTO_MIGRATE", "false") == "true" {
		if err := DB.AutoMigrate(&models.User{}); err != nil {
			log.Println("AutoMigrate failed:", err)
		}
	}

	log.Println("Database connected")
}

func getEnvWithDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
