package main

import (
	"coworkingapp/handlers"
	"coworkingapp/models"

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
