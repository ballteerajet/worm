package utils

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword แปลงรหัสผ่านเป็น Hash
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash ตรวจสอบว่ารหัสผ่านตรงกับ Hash หรือไม่
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}