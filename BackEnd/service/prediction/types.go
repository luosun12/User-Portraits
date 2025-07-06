package prediction

import "time"

// PredictionRequest 预测请求结构
type PredictionRequest struct {
	UserID      uint   `json:"user_id"`
	StationID   uint   `json:"station_id"`
	CurrentTime string `json:"current_time"`
	// 预测未来时间点数量（每个时间点间隔1小时）
	PredictHours int `json:"predict_hours"`
}

// TimePoint 时间点预测结果
type TimePoint struct {
	Time     time.Time `json:"time"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		District  string  `json:"district"`
		City      string  `json:"city"`
	} `json:"location"`
	NetworkMetrics struct {
		Flow     uint `json:"flow"`
		Latency  uint `json:"latency"`
		ErrCount uint `json:"err_count"`
	} `json:"network_metrics"`
	Confidence float64 `json:"confidence"` // 预测置信度
}

// PredictionResponse 预测响应结构
type PredictionResponse struct {
	UserID         uint        `json:"user_id"`
	Predictions    []TimePoint `json:"predictions"` // 按时间顺序排列的预测结果
	LastUpdateTime time.Time   `json:"last_update_time"`
}

// TrainingRequest 训练请求结构
type TrainingRequest struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Epochs    int    `json:"epochs"`
	BatchSize int    `json:"batch_size"`
}

// TrainingResponse 训练响应结构
type TrainingResponse struct {
	Message string `json:"message"`
}

// ModelStatus 模型状态结构
type ModelStatus struct {
	LatestModel string `json:"latest_model"`
	ModelPath   string `json:"model_path"`
	Metrics     struct {
		TrainingLoss float64 `json:"training_loss"`
		ValidLoss    float64 `json:"valid_loss"`
		Accuracy     float64 `json:"accuracy"`
	} `json:"metrics"`
	LastTrainingTime time.Time `json:"last_training_time"`
}

// ServiceConfig ML服务配置
type ServiceConfig struct {
	Host string
	Port int
}
