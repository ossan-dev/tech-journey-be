package handlers

import (
	"net/http"
	"time"

	"coworkingapp/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetAllRooms(c *gin.Context) {
	var dayToBook time.Time
	var err error
	rawDayToBook := c.Query("day_to_book")
	if rawDayToBook != "" {
		dayToBook, err = time.Parse("2006-01-02", rawDayToBook)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.CoworkingErr{Code: models.DateWrongFormatErr, Message: err.Error()})
			return
		}
	}
	db := c.MustGet("DbKey").(*gorm.DB)
	rooms, err := models.GetRooms(db, dayToBook)
	if err != nil {
		coworkingErr := err.(models.CoworkingErr)
		c.JSON(coworkingErr.StatusCode, coworkingErr)
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func GetRoomById(c *gin.Context) {
	db := c.MustGet("DbKey").(*gorm.DB)
	room, err := models.GetRoomById(db, c.Param("id"))
	if err != nil {
		coworkingErr := err.(models.CoworkingErr)
		c.JSON(coworkingErr.StatusCode, coworkingErr)
		return
	}
	c.JSON(http.StatusOK, room)
}

func GetRoomPhotos(c *gin.Context) {
	db := c.MustGet("DbKey").(*gorm.DB)
	photos, err := models.GetRoomPhotos(db, c.Param("id"))
	if err != nil {
		coworkingErr := err.(models.CoworkingErr)
		c.JSON(coworkingErr.StatusCode, coworkingErr)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": c.Param("id"), "photos": photos})
}
