package models

import (
	"time"
	"gorm.io/gorm"
)

// User: เก็บข้อมูลผู้ใช้
type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" json:"username"`
	
	// เติม json:"-" เพื่อบอกว่าเวลาแปลงเป็น JSON ไม่ต้องเอาค่านี้ออกไป
	Password string `gorm:"not null" json:"-"` 
	
	Role     string `gorm:"default:user" json:"role"`
	APIKey   string `gorm:"unique;index" json:"api_key"`
}

// SensorData: เก็บข้อมูลสภาพอากาศ
type SensorData struct {
	ID          uint      `gorm:"primaryKey"`
	Temperature float64   `gorm:"not null"`
	Humidity    float64   `gorm:"not null"`
	CreatedAt   time.Time // Auto timestamp
}

