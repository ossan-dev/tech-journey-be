package main

import (
	"coworkingapp/handlers"
	"coworkingapp/models"
	"coworkingapp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	gin.SetMode(gin.DebugMode)
	dsn := "host=localhost port=54322 user=postgres password=postgres dbname=postgres sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Room{})
	db.AutoMigrate(&models.Photo{})
	db.AutoMigrate(&models.Booking{})
	seedData(db)
	r := gin.Default()

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
