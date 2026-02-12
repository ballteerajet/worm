package models

import (
	"time"
	"gorm.io/gorm"
)

// User: เก็บข้อมูลผู้ใช้
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Role     string `gorm:"default:user"` // admin, user
	APIKey   string `gorm:"unique;index"`
}

// SensorData: เก็บข้อมูลสภาพอากาศ
type SensorData struct {
	ID          uint      `gorm:"primaryKey"`
	Temperature float64   `gorm:"not null"`
	Humidity    float64   `gorm:"not null"`
	CreatedAt   time.Time // Auto timestamp
}