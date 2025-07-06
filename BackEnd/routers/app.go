package main

import (
	"UserPortrait/configs"
	"UserPortrait/middleware"
	"UserPortrait/parsePacket/capture"
	"UserPortrait/parsePacket/process"
	"UserPortrait/service"
	"UserPortrait/service/prediction"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var predictionClient *prediction.Client

func InitRouter() *gin.Engine {
	r := gin.Default()
	CORS := cors.Config{
		AllowOrigins: []string{
			configs.CORS_ORIGIN,
		},
	}
	r.Use(cors.New(CORS))

	// 初始化预测服务客户端
	predictionClient = prediction.NewClient(prediction.ServiceConfig{
		Host: "localhost",
		Port: 8000,
	})

	public := r.Group("/public")
	{
		public.POST("/register", service.Register)
		public.POST("/login", service.Login)
		public.GET("/getUserBasicInfo", service.GetUserBasicInfo)

		public.POST("/admin_register", service.AdminRegister)
		public.POST("/admin_login", service.AdminLogin)
		public.GET("/getPrediction", service.GetPrediction)
		public.POST("/triggerTraining", service.TriggerTraining)
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
		// 添加预测接口
		// us.GET("/getPrediction", service.GetPrediction)
		sc := private.Group("/score")
		sc.GET("/average_score", service.GetAverageScore)

		ad := private.Group("/admin")
		ad.Use(middleware.AdminJwtAuthentication())
		ad.GET("/getStationInfo", service.GetBaseStationInfo)
		// 添加手动触发训练接口
		// ad.POST("/triggerTraining", service.TriggerTraining)
	}
	return r
}

// 定期训练任务
func scheduledTraining() {
	ticker := time.NewTicker(24 * time.Hour) // 每24小时训练一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			now := time.Now()
			startDate := now.AddDate(0, -1, 0).Format("2006-01-02") // 使用过去一个月的数据
			endDate := now.Format("2006-01-02")

			req := prediction.TrainingRequest{
				StartDate: startDate,
				EndDate:   endDate,
				Epochs:    50,
				BatchSize: 32,
			}

			if err := predictionClient.Train(req); err != nil {
				fmt.Printf("Scheduled training failed: %v\n", err)
			} else {
				fmt.Printf("Scheduled training completed successfully at %v\n", now)
			}
		}
	}
}

func main() {
	// 启动HTTP服务
	go func() {
		r := InitRouter()
		err := r.Run("localhost:5000")
		if err != nil {
			err := fmt.Errorf("failed to run server: %v", err)
			panic(err)
		}
	}()

	// 启动数据捕获
	go process.CapturePackets()
	go capture.Tcpd()

	// 启动定期训练
	time.Sleep(10 * time.Second)
	go scheduledTraining()

	// 信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("收到退出信号，程序退出...")
}
