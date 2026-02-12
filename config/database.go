package config

import (
	"fmt"
	"log"
	"os"
	"worm/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDB() *gorm.DB {
	// ลองโหลด .env (ถ้ามี) แต่ถ้าไม่มี (บน Render) ก็ข้ามไป
	_ = godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set in environment variables")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	fmt.Println("Database Connected & Migrated!")
	db.AutoMigrate(&models.User{}, &models.SensorData{})

	return db
}