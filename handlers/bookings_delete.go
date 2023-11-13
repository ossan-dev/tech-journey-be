package handlers

import (
	"net/http"

	"coworkingapp/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func DeleteBooking(c *gin.Context) {
	userId := c.MustGet("UserIdKey").(string)
	db := c.MustGet("DbKey").(*gorm.DB)
	if err := models.DeleteBookingById(db, c.Param("id"), userId); err != nil {
		coworkingErr := err.(models.CoworkingErr)
		c.JSON(coworkingErr.StatusCode, coworkingErr)
		return
	}
	c.Status(http.StatusNoContent)
}
