package handlers

import (
	"net/http"

	"coworkingapp/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetBookingsByUserId(c *gin.Context) {
	userId := c.MustGet("UserIdKey").(string)
	db := c.MustGet("DbKey").(*gorm.DB)
	bookings, err := models.GetBookingsByUserId(db, userId)
	if err != nil {
		coworkingErr := err.(models.CoworkingErr)
		c.JSON(coworkingErr.StatusCode, coworkingErr)
		return
	}
	c.JSON(http.StatusOK, bookings)
}

func GetBookingById(c *gin.Context) {
	userId := c.MustGet("UserIdKey").(string)
	db := c.MustGet("DbKey").(*gorm.DB)
	booking, err := models.GetBookingById(db, c.Param("id"), userId)
	if err != nil {
		coworkingErr := err.(models.CoworkingErr)
		c.JSON(coworkingErr.StatusCode, coworkingErr)
		return
	}
	c.JSON(http.StatusOK, booking)
}
