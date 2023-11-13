package main

import (
	"encoding/json"
	"os"

	"coworkingapp/handlers"
	"coworkingapp/middlewares"
	"coworkingapp/models"
	"coworkingapp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var config models.CoworkingConfig

func init() {
	data, err := os.ReadFile("config/config.json")
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data, &config); err != nil {
		panic(err)
	}
}

func main() {
	gin.SetMode(gin.DebugMode)
	db, err := gorm.Open(postgres.Open(config.Dsn))
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Room{})
	db.AutoMigrate(&models.Photo{})
	db.AutoMigrate(&models.Booking{})
	seedData(db)
	r := gin.Default()
	r.Use(middlewares.EarlyExitOnPreflightRequests())
	r.Use(middlewares.SetCorsPolicy(config.AllowedOrigins))
	r.Use(func(c *gin.Context) {
		c.Set("DbKey", db)
		c.Set("ConfigKey", config)
		c.Next()
	})
	r.POST("/auth/login", handlers.Login)
	r.POST("/auth/signup", handlers.Signup)
	r.GET("/rooms", handlers.GetAllRooms)
	r.GET("/rooms/:id", handlers.GetRoomById)
	r.GET("/rooms/:id/photos", handlers.GetRoomPhotos)
	r.GET("/bookings", handlers.GetBookingsByUserId)
	r.GET("/bookings/:id", handlers.GetBookingById)
	r.POST("/bookings", handlers.AddBooking)
	r.DELETE("/bookings/:id", handlers.DeleteBooking)

	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func seedData(db *gorm.DB) {
	db.Delete(&models.Booking{}, "1 = 1")
	db.Delete(&models.User{}, "1 = 1")
	db.Delete(&models.Photo{}, "1 = 1")
	db.Delete(&models.Room{}, "1 = 1")
	userId := utils.GetUuid()
	db.Create(&models.User{
		ID:       userId,
		Email:    "ipesenti@sorint.com",
		Username: "ipesenti",
		Password: "abcd1234!!",
	})
	db.Create([]*models.Room{
		{
			ID: utils.GetUuid(), Name: "Green", Cost: 12.50, NumberOfSeats: 4, Category: "Open Space", MainPhoto: "/green_0001.jpg", Photos: []models.Photo{
				{Url: "/green_0002.jpg"},
				{Url: "/green_0003.jpg"},
			},
		},
		{
			ID: utils.GetUuid(), Name: "Red", Cost: 100.00, NumberOfSeats: 50, Category: "Conference Hall", MainPhoto: "/red_0001.jpg", Photos: []models.Photo{
				{Url: "/red_0002.jpg"},
			},
		},
		{
			ID: utils.GetUuid(), Name: "Yellow", Cost: 4.50, NumberOfSeats: 1, Category: "Shared Desk", MainPhoto: "/yellow_0001.jpg", Photos: []models.Photo{
				{Url: "/yellow_0002.jpg"},
				{Url: "/yellow_0003.jpg"},
				{Url: "/yellow_0004.jpg"},
			},
		},
	})
}
