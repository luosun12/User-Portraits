package service

import (
	"UserPortrait/service/prediction"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

var predictionClient *prediction.Client

// InitPredictionClient 初始化预测客户端
func InitPredictionClient(config prediction.ServiceConfig) {
	predictionClient = prediction.NewClient(config)
}

// GetPrediction 获取预测结果
func GetPrediction(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 32)
	stationID, _ := strconv.ParseUint(c.Query("station_id"), 10, 32)
	predictHours, _ := strconv.Atoi(c.DefaultQuery("predict_hours", "24")) // 默认预测24小时

	if predictHours <= 0 || predictHours > 168 { // 最多预测一周
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "预测时间范围无效，应在1-168小时之间",
		})
		return
	}

	req := prediction.PredictionRequest{
		UserID:       uint(userID),
		StationID:    uint(stationID),
		CurrentTime:  time.Now().Format("2006-01-02 15:04:05"),
		PredictHours: predictHours,
	}

	// 使用预测客户端进行预测
	resp, err := predictionClient.Predict(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "预测失败",
			"error":   err.Error(),
		})
		return
	}

	// 处理预测结果
	result := gin.H{
		"message": "预测成功",
		"data": gin.H{
			"user_id":     resp.UserID,
			"predictions": resp.Predictions,
			"summary":     generatePredictionSummary(resp.Predictions),
			"last_update": resp.LastUpdateTime,
		},
	}

	c.JSON(http.StatusOK, result)
}

// generatePredictionSummary 生成预测结果摘要
func generatePredictionSummary(predictions []prediction.TimePoint) gin.H {
	var (
		totalFlow     uint = 0
		totalLatency  uint = 0
		totalErrCount uint = 0
		maxFlow       uint = 0
		peakTime      time.Time
	)

	for _, p := range predictions {
		totalFlow += p.NetworkMetrics.Flow
		totalLatency += p.NetworkMetrics.Latency
		totalErrCount += p.NetworkMetrics.ErrCount

		if p.NetworkMetrics.Flow > maxFlow {
			maxFlow = p.NetworkMetrics.Flow
			peakTime = p.Time
		}
	}

	count := len(predictions)
	if count == 0 {
		return gin.H{}
	}

	return gin.H{
		"average_flow":     totalFlow / uint(count),
		"average_latency":  totalLatency / uint(count),
		"total_errors":     totalErrCount,
		"peak_flow":        maxFlow,
		"peak_time":        peakTime,
		"prediction_count": count,
	}
}

// TriggerTraining 手动触发模型训练
func TriggerTraining(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		// 默认使用过去一个月的数据
		now := time.Now()
		endDate = now.Format("2006-01-02")
		startDate = now.AddDate(0, -1, 0).Format("2006-01-02")
	}

	req := prediction.TrainingRequest{
		StartDate: startDate,
		EndDate:   endDate,
		Epochs:    50,
		BatchSize: 32,
	}

	if err := predictionClient.Train(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "训练失败",
			"error":   err.Error(),
		})
		return
	}

	// 获取模型状态
	status, err := predictionClient.GetModelStatus()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "训练成功，但获取模型状态失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "训练成功",
		"model_status": status,
	})
}
