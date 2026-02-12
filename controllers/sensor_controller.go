package controllers

import (
	"worm/models"
	"gorm.io/gorm"
)

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