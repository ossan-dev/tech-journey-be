package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/ossan-dev/coworkingapp/middlewares"
)

func SetupRoutes(r *gin.Engine, fr *FlightRecorderTracer) {
	r.Static("/imgs", "./imgs")
	r.POST("/auth/login", Login)
	r.POST("/auth/signup", Signup)
	r.GET("/trace", fr.Trace)
	r.GET("/rooms", GetAllRooms)
	r.GET("/rooms/:id", GetRoomById)
	r.GET("/rooms/:id/photos", GetRoomPhotos)
	r.GET("/bookings", middlewares.AuthorizeUser(), GetBookingsByUserId)
	r.GET("/bookings/:id", middlewares.AuthorizeUser(), GetBookingById)
	r.POST("/bookings", middlewares.AuthorizeUser(), AddBooking)
	r.DELETE("/bookings/:id", middlewares.AuthorizeUser(), DeleteBooking)
}
