package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/ossan-dev/coworkingapp/models"
	"gorm.io/gorm"
)

//go:noinline
func SetRequestValues(db gorm.DB, config models.CoworkingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("DbKey", db)
		c.Set("ConfigKey", config)
		c.Next()
	}
}
