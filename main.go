package main

import (
	"coworkingapp/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.DebugMode)
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
