package middlewares

import (
	"net/http"

	"github.com/ossan-dev/coworkingapp/models"
	"github.com/ossan-dev/coworkingapp/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AuthorizeUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.GetHeader("Authorization")
		if tokenHeader == "" {
			c.JSON(http.StatusUnauthorized, models.CoworkingErr{Code: models.MissingTokenErr, Message: "please provide a jwt token along with the HTTP headers"})
			return
		}
		secretKey := c.MustGet("ConfigKey").(models.CoworkingConfig).SecretKey
		claims, err := utils.ValidateToken(tokenHeader, []byte(secretKey))
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.CoworkingErr{Code: models.TokenNotValidErr, Message: err.Error()})
			return
		}
		email := (*claims)["sub"].(string)
		db := c.MustGet("DbKey").(*gorm.DB)
		user, err := models.GetUserByEmail(db, email)
		if err != nil {
			coworkingErr := err.(models.CoworkingErr)
			c.JSON(coworkingErr.StatusCode, coworkingErr)
			return
		}
		c.Set("UserIdKey", user.ID)
		c.Next()
	}
}
