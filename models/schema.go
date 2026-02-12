package models

import (
	"time"
	"gorm.io/gorm"
)

// User: เก็บข้อมูลผู้ใช้
type User struct {
	// ลบ gorm.Model ทิ้ง แล้วใส่ 3 บรรทัดนี้แทน
	ID        uint           `gorm:"primaryKey" json:"id" example:"1"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // json:"-" คือซ่อน field นี้ไม่ให้โชว์ใน Docs/Response

	// Fields เดิมของคุณ
	Username string `gorm:"unique;not null" json:"username" example:"staff01"`
	Password string `gorm:"not null" json:"-"` 
	Role     string `gorm:"default:user" json:"role" example:"user"`
	APIKey   string `gorm:"unique;index" json:"api_key"`
}

// SensorData: เก็บข้อมูลสภาพอากาศ
type SensorData struct {
	// ทำเหมือนกัน
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	// SensorData อาจจะไม่จำเป็นต้องมี UpdatedAt/DeletedAt ก็ได้แล้วแต่ design
	
	Temperature float64 `gorm:"not null" json:"temp" example:"32.5"`
	Humidity    float64 `gorm:"not null" json:"humidity" example:"60.0"`
}