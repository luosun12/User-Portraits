package middleware

import (
	"UserPortrait/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 登录状态token验证中间件
func UserJwtAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.UserTokenValid(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		}
		c.Next()
	}
}

func AdminJwtAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.AdminTokenValid(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		}
		c.Next()
	}
}
