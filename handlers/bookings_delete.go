package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func DeleteBooking(c *gin.Context) {
	c.String(http.StatusOK, "DeleteBooking()")
}
