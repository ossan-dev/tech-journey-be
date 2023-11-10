package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func SetCorsPolicy(allowedOrigins string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", allowedOrigins)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")
		c.Next()
	}
}

func EarlyExitOnPreflightRequests() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusOK)
			return
		}
		c.Next()
	}
}
