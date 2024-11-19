package main

import (
	"UserPortrait/configs"
	"UserPortrait/middleware"
	"UserPortrait/parsePacket/capture"
	"UserPortrait/parsePacket/process"
	"UserPortrait/service"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"os"
	"os/signal"
	"syscall"
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
		public.POST("/register", service.Register)
		public.POST("/login", service.Login)
		public.GET("/getUserBasicInfo", service.GetUserBasicInfo)

		public.POST("/admin_register", service.AdminRegister)
		public.POST("/admin_login", service.AdminLogin)
	}
	private := r.Group("")
	{
		// TODO:主页面请求内容，暂用Ping替代
		private.GET("/main", service.Ping)

		us := private.Group("/user")
		us.Use(middleware.UserJwtAuthentication())
		us.POST("/avatar", service.UploadAvatar)
		us.POST("/score", service.SubmitScore)
		us.POST("/reset_password", service.ResetPassword)
		us.GET("/getDailyFlow", service.GetUserDailyFlow)
		us.GET("/getFrequentPlaces", service.GetFreqLocation)

		sc := private.Group("/score")
		sc.GET("/average_score", service.GetAverageScore)

		ad := private.Group("/admin")
		ad.Use(middleware.AdminJwtAuthentication())
		ad.GET("/getStationInfo", service.GetBaseStationInfo)
	}
	return r
}

func main() {
	go func() {
		r := InitRouter()
		err := r.Run("localhost:5000")
		if err != nil {
			err := fmt.Errorf("failed to run server: %v", err)
			panic(err)
		}
	}()
	go process.CapturePackets()
	go capture.Tcpd()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("收到退出信号，程序退出...")
}
