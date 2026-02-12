package controllers

import (
	"net/http"
	"worm/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// --- 1. Request Models (สำหรับ Swagger) ---

// SensorRequest แบบฟอร์มรับค่า Sensor
type SensorRequest struct {
	Temperature float64 `json:"temp" example:"32.5"`
	Humidity    float64 `json:"humidity" example:"60.0"`
}

// --- 2. Handlers (สำหรับ Gin & Swagger) ---

// AddSensorHandler บันทึกค่า Sensor
// @Summary      บันทึกค่าอุณหภูมิและความชื้น
// @Description  รับค่า temp และ humidity แล้วบันทึกลงฐานข้อมูล
// @Tags         Sensor
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        request body SensorRequest true "ข้อมูล Sensor"
// @Success      200  {object} map[string]string "message: Saved"
// @Failure      400  {object} map[string]string "error: Bad Request"
// @Router       /sensor [post]
func AddSensorHandler(c *gin.Context, db *gorm.DB) {
	var req SensorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// เรียก Logic ภายใน
	if err := AddSensorData(db, req.Temperature, req.Humidity); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Saved"})
}

// GetAllSensorHandler ดูข้อมูล Sensor ทั้งหมด
// @Summary      ดูประวัติ Sensor
// @Description  ดึงข้อมูลอุณหภูมิและความชื้นทั้งหมด
// @Tags         Sensor
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200  {array} models.SensorData
// @Router       /sensor [get]
func GetAllSensorHandler(c *gin.Context, db *gorm.DB) {
	data, err := GetAllSensorData(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// --- 3. Internal Logic (Functions เดิมของคุณ) ---

func AddSensorData(db *gorm.DB, temp float64, humidity float64) error {
	data := models.SensorData{
		Temperature: temp,
		Humidity:    humidity,
	}
	return db.Create(&data).Error
}

func GetAllSensorData(db *gorm.DB) ([]models.SensorData, error) {
	var data []models.SensorData
	err := db.Find(&data).Error
	return data, err
}