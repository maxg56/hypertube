package conf

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

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

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	const maxRetries = 10
	var (
		database *gorm.DB
		err      error
	)
	for i := range maxRetries {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
		if err == nil {
			database = database.Session(&gorm.Session{Logger: gormLogger})
			break
		}
		log.Printf("Database not ready, retrying in 2s (%d/%d): %v", i+1, maxRetries, err)
		time.Sleep(2 * time.Second)
	}
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
			&models.OAuthAccount{},
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
