package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAllRooms(c *gin.Context) {
	c.String(http.StatusOK, "GetAllRooms()")
}

func GetRoomById(c *gin.Context) {
	c.String(http.StatusOK, "GetRoomById()")
}

func GetRoomPhotos(c *gin.Context) {
	c.String(http.StatusOK, "GetRoomPhotos()")
}
