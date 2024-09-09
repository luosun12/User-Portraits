package middleware

import (
	"UserPortrait/token"
	"github.com/gin-gonic/gin"
	"net/http"
)

// 登录状态token验证中间件
func JwtAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := token.TokenValid(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
		}
		c.Next()
	}
}
