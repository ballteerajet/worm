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

	// Import Swagger
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Import Docs ที่จะถูกสร้าง (เปลี่ยน 'worm' เป็นชื่อ module ของคุณถ้าไม่ใช่)
	_ "worm/docs"
)

// --- Structs สำหรับ Swagger (ต้องประกาศข้างนอกเพื่อให้ Swagger เห็น) ---

// LoginRequest โมเดลสำหรับ Login
type LoginRequest struct {
	Username string `json:"username" example:"root_admin" binding:"required"`
	Password string `json:"password" example:"admin1234" binding:"required"`
}

// RegisterRequest โมเดลรับข้อมูลสมัครสมาชิก
type RegisterRequest struct {
	Username string `json:"username" example:"staff01" binding:"required"`
	Password string `json:"password" example:"123456" binding:"required"`
	Role     string `json:"role" example:"user" binding:"required"` // admin หรือ user
}

// SensorRequest โมเดลรับข้อมูล Sensor
type SensorRequest struct {
	Temp     float64 `json:"temp" example:"32.5" binding:"required"`
	Humidity float64 `json:"humidity" example:"60.0" binding:"required"`
}

// --- Main Setup ---

// @title           IoT Sensor API
// @version         1.0
// @description     ระบบ API สำหรับจัดการ User และข้อมูล Sensor
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.email   support@swagger.io

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            worm-bwqp.onrender.com
// @BasePath        /api

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-KEY
func main() {
	db := config.ConnectDB()

	// --- Auto Create First Admin ---
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count == 0 {
		hashedPw, _ := utils.HashPassword("admin1234")
		firstAdmin := models.User{
			Username: "root_admin",
			Password: hashedPw,
			Role:     "admin",
			APIKey:   uuid.New().String(),
		}
		db.Create(&firstAdmin)
		fmt.Println("!!! FIRST ADMIN CREATED !!! Key:", firstAdmin.APIKey)
	}

	r := gin.Default()

	// LoginHandler
	// @Summary      เข้าสู่ระบบ (Login)
	// @Description  ส่ง Username/Password เพื่อรับ API Key
	// @Tags         Auth
	// @Accept       json
	// @Produce      json
	// @Param        request body LoginRequest true "Login Credentials"
	// @Success      200  {object} map[string]string
	// @Failure      401  {object} map[string]string
	// @Router       /login [post]
	r.POST("/login", func(c *gin.Context) {
		var req LoginRequest
		
		// 1. รับค่า JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
			return
		}

		// 2. ค้นหา User ใน Database
		var user models.User
		if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// 3. ตรวจสอบรหัสผ่าน (ใช้ฟังก์ชันจาก utils ที่เราเขียนไว้)
		if !utils.CheckPasswordHash(req.Password, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect password"})
			return
		}

		// 4. ถ้าผ่านหมด ส่ง API Key กลับไป
		c.JSON(http.StatusOK, gin.H{
			"message":  "Login successful",
			"username": user.Username,
			"role":     user.Role,
			"api_key":  user.APIKey, // <--- พระเอกของเราอยู่นี่
		})
	})

	// --- Route สำหรับ Swagger ---
	// เข้าผ่าน: /swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Redirect หน้าแรกไป Swagger เลย
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	protected := r.Group("/api")
	protected.Use(func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-KEY")
		user, err := middleware.Authenticate(db, apiKey, "user")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Set("user", user)
		c.Next()
	})

	// ---------------- API ROUTES ----------------

	// RegisterHandler
	// @Summary      สร้าง User ใหม่
	// @Description  สร้าง User หรือ Admin (เฉพาะ Admin เท่านั้นที่ใช้ได้)
	// @Tags         Auth
	// @Accept       json
	// @Produce      json
	// @Security     ApiKeyAuth
	// @Param        request body RegisterRequest true "User Info"
	// @Success      200  {object} map[string]interface{}
	// @Failure      403  {object} map[string]string
	// @Router       /register [post]
	protected.POST("/register", func(c *gin.Context) {
		requester := c.MustGet("user").(*models.User)
		if requester.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}

		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Role != "admin" && req.Role != "user" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role must be 'admin' or 'user'"})
			return
		}

		hashedPw, _ := utils.HashPassword(req.Password)
		newUser := models.User{
			Username: req.Username,
			Password: hashedPw,
			Role:     req.Role,
			APIKey:   uuid.New().String(),
		}

		if err := db.Create(&newUser).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Username taken"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Created", "user": newUser})
	})

	// GetUsersHandler
	// @Summary      ดูรายชื่อ User ทั้งหมด
	// @Description  ดึงข้อมูล User ทั้งหมดในระบบ (เฉพาะ Admin)
	// @Tags         Auth
	// @Produce      json
	// @Security     ApiKeyAuth
	// @Success      200  {object} map[string]interface{}
	// @Router       /users [get]
	protected.GET("/users", func(c *gin.Context) {
		requester := c.MustGet("user").(*models.User)
		if requester.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin only"})
			return
		}
		users, _ := controllers.GetAllUsers(db)
		c.JSON(http.StatusOK, gin.H{"users": users})
	})

	// AddSensorHandler
	// @Summary      ส่งค่า Sensor
	// @Description  บันทึกค่าอุณหภูมิและความชื้น
	// @Tags         Sensor
	// @Accept       json
	// @Produce      json
	// @Security     ApiKeyAuth
	// @Param        request body SensorRequest true "Sensor Data"
	// @Success      200  {object} map[string]string
	// @Router       /sensor [post]
	protected.POST("/sensor", func(c *gin.Context) {
		var req SensorRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		controllers.AddSensorData(db, req.Temp, req.Humidity)
		c.JSON(http.StatusOK, gin.H{"message": "Saved"})
	})

	// GetSensorHandler
	// @Summary      ดูข้อมูล Sensor
	// @Description  ดูประวัติค่าอุณหภูมิทั้งหมด
	// @Tags         Sensor
	// @Produce      json
	// @Security     ApiKeyAuth
	// @Success      200  {object} map[string]interface{}
	// @Router       /sensor [get]
	protected.GET("/sensor", func(c *gin.Context) {
		data, _ := controllers.GetAllSensorData(db)
		c.JSON(http.StatusOK, gin.H{"data": data})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "0000"
	}
	r.Run(":" + port)
}