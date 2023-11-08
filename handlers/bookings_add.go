package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddBooking(c *gin.Context) {
	c.String(http.StatusOK, "AddBooking()")
}
