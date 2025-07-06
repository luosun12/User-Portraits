package prediction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	config     ServiceConfig
	httpClient *http.Client
}

// NewClient 创建新的预测服务客户端
func NewClient(config ServiceConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
}

// Predict 调用预测接口
func (c *Client) Predict(req PredictionRequest) (*PredictionResponse, error) {
	url := fmt.Sprintf("http://%s:%d/predict", c.config.Host, c.config.Port)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %v", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("prediction request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("prediction service returned status: %d", resp.StatusCode)
	}

	var result PredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response failed: %v", err)
	}

	return &result, nil
}

// Train 调用训练接口
func (c *Client) Train(req TrainingRequest) error {
	url := fmt.Sprintf("http://%s:%d/train", c.config.Host, c.config.Port)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("marshal request failed: %v", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("training request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("training service returned status: %d", resp.StatusCode)
	}

	return nil
}

// GetModelStatus 获取模型状态
func (c *Client) GetModelStatus() (*ModelStatus, error) {
	url := fmt.Sprintf("http://%s:%d/model/status", c.config.Host, c.config.Port)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("status request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status service returned status: %d", resp.StatusCode)
	}

	var status ModelStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decode status failed: %v", err)
	}

	return &status, nil
}
