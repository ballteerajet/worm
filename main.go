package main

import (
	"fmt"
	"net/http"
	"os"

	"worm/config"
	"worm/controllers"
	"worm/middleware"
	"worm/models"
	"worm/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	// Import Swagger Files
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Import Docs ที่ Gen แล้ว (สำคัญมาก)
	_ "worm/docs"
)

// --- Structs สำหรับ Swagger (เฉพาะ Login ที่ยังไม่อยู่ใน Controller) ---

// LoginRequest โมเดลสำหรับ Login
type LoginRequest struct {
	Username string `json:"username" example:"root_admin" binding:"required"`
	Password string `json:"password" example:"admin1234" binding:"required"`
}

func main() {
	// 1. เชื่อมต่อฐานข้อมูล
	db := config.ConnectDB()

	// 2. ระบบ Auto Create First Admin (ถ้าไม่มี Admin เลย)
	var adminCount int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&adminCount)

	if adminCount == 0 {
		hashedPw, _ := utils.HashPassword("admin1234")
		firstAdmin := models.User{
			Username: "root_admin",
			Password: hashedPw,
			Role:     "admin",
			APIKey:   uuid.New().String(),
		}
		db.Create(&firstAdmin)
		fmt.Println("==================================================")
		fmt.Println("⚠️  NO ADMIN FOUND -> CREATED ROOT ADMIN")
		fmt.Printf("Username: %s\n", firstAdmin.Username)
		fmt.Printf("Password: admin1234\n")
		fmt.Printf("API Key:  %s\n", firstAdmin.APIKey)
		fmt.Println("==================================================")
	}

	// 3. เริ่มต้น Router
	r := gin.Default()

	// --- Public Routes ---

	// Swagger Route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Login Route (เรียกใช้ Function ด้านล่าง)
	r.POST("/login", func(c *gin.Context) {
		LoginHandler(c, db)
	})

	// --- Protected Routes (ต้องมี API Key) ---
	protected := r.Group("/api")
	protected.Use(func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		// เรียก Middleware เช็คสิทธิ์ (ขั้นต่ำคือ user)
		user, err := middleware.Authenticate(db, apiKey, "user")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Set("user", user)
		c.Next()
	})

	// 1. Register (Admin Only)
	protected.POST("/register", func(c *gin.Context) {
		// เช็คสิทธิ์ Admin ก่อนเรียก Controller
		requester := c.MustGet("user").(*models.User)
		if requester.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}
		// เรียกใช้ Handler จากไฟล์ controllers/user_controller.go
		controllers.CreateUserHandler(c, db)
	})

	// 2. Get Users (Admin Only)
	protected.GET("/users", func(c *gin.Context) {
		requester := c.MustGet("user").(*models.User)
		if requester.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}
		controllers.GetAllUsersHandler(c, db)
	})

	// 3. Sensor POST (All Users)
	protected.POST("/sensor", func(c *gin.Context) {
		controllers.AddSensorHandler(c, db)
	})

	// 4. Sensor GET (All Users)
	protected.GET("/sensor", func(c *gin.Context) {
		controllers.GetAllSensorHandler(c, db)
	})

	// Run Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

func LoginHandler(c *gin.Context, db *gorm.DB) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	var user models.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Login successful",
		"username": user.Username,
		"role":     user.Role,
		"api_key":  user.APIKey,
	})
}