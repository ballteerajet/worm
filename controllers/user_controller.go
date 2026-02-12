package controllers

import (
	"worm/models"
	"worm/utils"
	"worm/middleware"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateUser ฟังก์ชันสร้าง User ใหม่ (ทำได้เฉพาะ Admin)
func CreateUser(db *gorm.DB, adminAPIKey string, newUsername, newPassword, role string) (string, error) {
	// 1. เช็คสิทธิ์คนเรียกใช้ฟังก์ชันนี้ (ต้องเป็น Admin เท่านั้น)
	_, err := middleware.Authenticate(db, adminAPIKey, "admin") // *แก้ import cycle โดยย้าย logic หรือเรียกใช้ข้าม package ให้ถูก (ดู note ด้านล่าง)
	if err != nil {
		return "", err
	}

	// 2. Hash Password
	hashedPassword, _ := utils.HashPassword(newPassword)

	// 3. สร้าง User + Generate API Key
	generatedKey := uuid.New().String()
	newUser := models.User{
		Username: newUsername,
		Password: hashedPassword,
		Role:     role,
		APIKey:   generatedKey,
	}

	if err := db.Create(&newUser).Error; err != nil {
		return "", err
	}

	return generatedKey, nil
}