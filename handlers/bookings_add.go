package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type BookingDto struct {
	RoomId   string `json:"room_id" binding:"required"`
	BookedOn string `json:"booked_on" binding:"required"`
}

func AddBooking(c *gin.Context) {
	var bookingDto BookingDto
	if err := c.ShouldBind(&bookingDto); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, bookingDto)
}
