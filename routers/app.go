package main

import (
	"UserPortrait/middleware"
	"UserPortrait/service"
	"fmt"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	public := r.Group("/public")
	{
		public.GET("/ping", service.Ping)
		public.POST("/register", service.Register)
		public.POST("/login", service.Login)
		public.GET("/getbasicinfo/:id", service.GetBasicInfo)
	}
	private := r.Group("")
	{
		// 使用gin自定义中间件
		private.Use(middleware.JwtAuthentication())
		private.GET("/main", service.Ping)
		private.POST("/upload/avatar", service.UploadAvatar)
	}
	return r
}

func main() {
	r := InitRouter()
	err := r.Run("localhost:5000")
	if err != nil {
		err := fmt.Errorf("failed to run server: %v", err)
		panic(err)
	}
}
