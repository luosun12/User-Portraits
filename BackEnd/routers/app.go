package main

import (
	"UserPortrait/configs"
	"UserPortrait/middleware"
	"UserPortrait/service"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	r := gin.Default()
	CORS := cors.Config{
		AllowOrigins: []string{
			configs.CORS_ORIGIN,
		},
	}
	r.Use(cors.New(CORS))
	public := r.Group("/public")
	{
		public.GET("/ping", service.Ping)
		public.POST("/register", service.Register)
		public.POST("/login", service.Login)
		public.GET("/getbasicinfo", service.GetBasicInfo)
		public.GET("/getstationinfo", service.GetBaseStationInfo)
	}
	private := r.Group("")
	{
		// 使用gin自定义中间件
		private.Use(middleware.JwtAuthentication())
		private.GET("/main", service.Ping)

		up := private.Group("/upload")
		up.POST("/avatar", service.UploadAvatar)
		up.POST("/score", service.SubmitScore)

		sc := private.Group("/score")
		sc.GET("/average_score", service.GetAverageScore)
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
