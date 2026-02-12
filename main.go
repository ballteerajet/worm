package main

import (
	"net/http"
	"os"

	"worm/config"     // เปลี่ยนเป็นชื่อ module คุณ
	"worm/controllers"
	"worm/middleware"
	"worm/models"
	"worm/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// 1. เชื่อมต่อ DB
	db := config.ConnectDB()

	// 2. เริ่มต้น Web Server (Gin)
	r := gin.Default()

	// --- Route สำหรับ Login / Register (ไม่ต้องใช้ Middleware) ---
	r.POST("/register", func(c *gin.Context) {
		var json struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// สร้าง User ใหม่
		hashedPw, _ := utils.HashPassword(json.Password)
		newUser := models.User{
			Username: json.Username,
			Password: hashedPw,
			Role:     "user",
			APIKey:   uuid.New().String(),
		}

		if err := db.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Username might already exist"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User created", "api_key": newUser.APIKey})
	})

	// --- Route ที่ต้องใช้ API Key (Protected) ---
	protected := r.Group("/api")
	protected.Use(func(c *gin.Context) {
		// Middleware แบบบ้านๆ สำหรับ Gin
		apiKey := c.GetHeader("X-API-KEY") // รับ Key จาก Header
		user, err := middleware.Authenticate(db, apiKey, "user")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		// เก็บ user ไว้ใน context เผื่อใช้ต่อ
		c.Set("user", user)
		c.Next()
	})

	// POST: ส่งข้อมูล Sensor
	protected.POST("/sensor", func(c *gin.Context) {
		var json struct {
			Temp     float64 `json:"temp"`
			Humidity float64 `json:"humidity"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		controllers.AddSensorData(db, json.Temp, json.Humidity)
		c.JSON(http.StatusOK, gin.H{"message": "Data saved"})
	})

	// GET: ดูข้อมูล Sensor
	protected.GET("/sensor", func(c *gin.Context) {
		data, _ := controllers.GetAllSensorData(db)
		c.JSON(http.StatusOK, gin.H{"data": data})
	})

	// 3. รัน Server (รองรับ Port ของ Render)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default สำหรับรันในเครื่อง
	}
	r.Run(":" + port)
}