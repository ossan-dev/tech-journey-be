package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetBookingsByUserId(c *gin.Context) {
	c.String(http.StatusOK, "GetBookingsByUserId()")
}

func GetBookingById(c *gin.Context) {
	c.String(http.StatusOK, "GetBookingById()")
}
