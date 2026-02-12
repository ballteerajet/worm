package middleware

import (
	"errors"
	"worm/models"
	"gorm.io/gorm"
)

// Authenticate ตรวจสอบ API Key และ Role
func Authenticate(db *gorm.DB, apiKey string, requiredRole string) (*models.User, error) {
	var user models.User
	
	// 1. เช็คว่ามี API Key นี้ในระบบไหม
	result := db.Where("api_key = ?", apiKey).First(&user)
	if result.Error != nil {
		return nil, errors.New("Unauthorized: Invalid API Key")
	}

	// 2. เช็ค Role (ถ้าต้องการ admin แต่ user เป็นแค่ user ธรรมดา -> error)
	if requiredRole == "admin" && user.Role != "admin" {
		return nil, errors.New("Forbidden: Admin access required")
	}

	// ผ่านทุกขั้นตอน Return user กลับไป
	return &user, nil
}