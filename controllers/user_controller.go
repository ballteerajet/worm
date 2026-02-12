package controllers

import (
	"net/http"
	"worm/models"
	"worm/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- 1. Request Models ---

// RegisterRequest แบบฟอร์มสมัครสมาชิก
type RegisterRequest struct {
	Username string `json:"username" example:"staff01" binding:"required"`
	Password string `json:"password" example:"pass1234" binding:"required"`
	Role     string `json:"role" example:"user" binding:"required"` // admin หรือ user
}

// --- 2. Handlers ---

// CreateUserHandler สร้าง User ใหม่
// @Summary      สร้าง User ใหม่ (Admin Only)
// @Description  สร้าง User หรือ Admin ใหม่ โดยต้องใช้ Key ของ Admin
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body RegisterRequest true "ข้อมูล User ใหม่"
// @Success      200  {object} map[string]interface{}
// @Failure      400  {object} map[string]string
// @Failure      500  {object} map[string]string
// @Router       /register [post]
func CreateUserHandler(c *gin.Context, db *gorm.DB) {
	// (สมมติว่า Middleware เช็คสิทธิ์ Admin ให้แล้วที่ main.go)

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := CreateUser(db, req.Username, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user (username might exist)"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created",
		"username": req.Username,
		"api_key": apiKey,
	})
}

// UpdateUserRequest: โมเดลสำหรับรับค่าแก้ไข (ใส่ * เพื่อเช็คว่าส่งมาหรือไม่)
type UpdateUserRequest struct {
	Username *string `json:"username" example:"new_name"`
	Password *string `json:"password" example:"new_pass123"`
	Role     *string `json:"role" example:"admin"`
}

// UpdateUserHandler แก้ไขข้อมูล User
// @Summary      แก้ไขข้อมูล User (Admin Only)
// @Description  อัปเดต Username, Password หรือ Role (ส่งเฉพาะค่าที่ต้องการแก้)
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id      path   int                true  "User ID"
// @Param        request body   UpdateUserRequest  true  "ข้อมูลที่ต้องการแก้"
// @Success      200     {object} models.User
// @Failure      400     {object} map[string]string
// @Failure      404     {object} map[string]string
// @Router       /users/{id} [put]
func UpdateUserHandler(c *gin.Context, db *gorm.DB) {
	// 1. รับ ID
	id := c.Param("id")

	// 2. หา User เดิม
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 3. รับค่าจาก Body
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 4. อัปเดตทีละค่า (ถ้าส่งมา)
	if req.Username != nil {
		user.Username = *req.Username
	}
	if req.Role != nil {
		if *req.Role == "admin" || *req.Role == "user" {
			user.Role = *req.Role
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be 'admin' or 'user'"})
			return
		}
	}
	if req.Password != nil {
		hashed, _ := utils.HashPassword(*req.Password)
		user.Password = hashed
	}

	// 5. บันทึก
	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed (username might exist)"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAllUsersHandler ดูรายชื่อ User
// @Summary      ดูรายชื่อ User ทั้งหมด
// @Description  แสดงรายชื่อ User และ Role ทั้งหมด (Admin Only)
// @Tags         Auth
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array} models.User
// @Router       /users [get]
func GetAllUsersHandler(c *gin.Context, db *gorm.DB) {
	users, err := GetAllUsers(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// --- 3. Internal Logic ---

// CreateUser (Logic ล้วน - ตัด Middleware ออกไปเช็คที่ Router แทน)
func CreateUser(db *gorm.DB, newUsername, newPassword, role string) (string, error) {
	hashedPassword, _ := utils.HashPassword(newPassword)
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

func GetAllUsers(db *gorm.DB) ([]models.User, error) {
	var users []models.User
	result := db.Find(&users)
	return users, result.Error
}