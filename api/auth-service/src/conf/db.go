package conf

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	models "auth-service/src/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	host := GetenvOrDefault("DB_HOST", "localhost")
	port := GetenvOrDefault("DB_PORT", "5432")
	user := GetenvOrDefault("DB_USER", "postgres")
	password := GetenvOrDefault("DB_PASSWORD", "password")
	dbname := GetenvOrDefault("DB_NAME", "hypertube")

	dsn := "host=" + host + " user=" + user + " password=" + password +
		" dbname=" + dbname + " port=" + port + " sslmode=disable"

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = database
	log.Println("Database connected")

	if os.Getenv("AUTO_MIGRATE") == "true" {
		if err := DB.AutoMigrate(
			&models.Users{},
			&models.EmailVerification{},
			&models.PasswordReset{},
		); err != nil {
			log.Println("AutoMigrate failed:", err)
		}
	}
}

func GetenvOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
