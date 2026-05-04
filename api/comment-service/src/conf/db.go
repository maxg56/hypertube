package conf

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() error {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbEnv("DB_HOST", "localhost"),
		dbEnv("DB_PORT", "5432"),
		dbEnv("DB_USER", "postgres"),
		dbEnv("DB_PASSWORD", "password"),
		dbEnv("DB_NAME", "hypertube"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	log.Println("Database connected")
	return nil
}

func dbEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
