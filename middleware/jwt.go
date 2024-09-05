package middleware

import (
	"UserPortrait/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

func JwtAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.TokenValid(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		}
		c.Next()
	}
}
