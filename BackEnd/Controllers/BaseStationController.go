package Controllers

import (
	"UserPortrait/etc"
	"UserPortrait/functions"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"sync"
)

var stationMutex sync.Mutex

// UpdateStationAfterUni 在Universe之后更新相应基站记录
func (s *SqlController) UpdateStationAfterUni(TableName string) error {
	stationMutex.Lock()
	defer stationMutex.Unlock()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		newStationRecord := <-etc.StationChannel
		record, err := s.FindStationRecordByTime(TableName, newStationRecord.Date, newStationRecord.PeriodID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 若找不到记录，插入新记录
			err = s.DB.Table(TableName).Create(&newStationRecord).Error
			if err != nil {
				return fmt.Errorf("failed to create new station record: %v", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to create new station record: %v", err)
		} else {
			// 若错误，则新记录errcount默认为1
			errCount := record.ErrCount + newStationRecord.ErrCount
			totalFlow := newStationRecord.TotalFlow + record.TotalFlow
			aveLatency := (newStationRecord.AveLatency + record.AveLatency*record.ConnCount) / (record.ConnCount + 1)
			err = s.DB.Table(TableName).Where("date =? AND period_id =?", record.Date, record.PeriodID).Updates(map[string]interface{}{
				"conn_count":  record.ConnCount + 1,
				"err_count":   errCount,
				"total_flow":  totalFlow,
				"ave_latency": aveLatency,
			}).Error
			return err
		}
		return nil
	})
}

func (s *SqlController) FindStationRecordByTime(TableName string, date string, periodID uint) (etc.BaseStation, error) {
	var record etc.BaseStation
	FoundRecord := s.DB.Table(TableName).Where("date =? AND period_id =?", date, periodID).Take(&record)
	return record, FoundRecord.Error
}

// DailyStationRecords 管理员用：获取指定基站的近24小时性能数据
func (s *SqlController) DailyStationRecords(stationId uint, TableName string, yesterday string, today string, lastID uint, currID uint) (etc.StationInterface, error) {
	var lastRecords []etc.BaseStation
	var currRecords []etc.BaseStation
	// 实例初始化
	var entity etc.StationInterface
	// 更新实例的静态信息
	entity.StationInfo.StationID = stationId
	entity.StationInfo.Latitute, entity.StationInfo.Longitude = functions.ChooseStationLoc(stationId)
	entity.CurrentPeriod = currID

	// 数据库条件遍历，获取昨日、今日近24小时记录
	Contents1 := s.DB.Table(TableName).Select("*")
	Contents2 := s.DB.Table(TableName).Select("*")
	err1 := Contents1.Where("date =? AND period_id >=? AND period_id <=?", yesterday, int(lastID), 24).Order("period_id").Find(&lastRecords).Error
	if err1 != nil {
		return entity, err1
	}
	err2 := Contents2.Where("date =? AND period_id >=? AND period_id <=?", today, 1, int(currID)).Order("period_id").Find(&currRecords).Error
	if err2 != nil {
		return entity, err2
	}

	for _, record1 := range currRecords {
		entity.Status = append(entity.Status, struct {
			PeriodID        uint    `json:"time_id"`
			ConnCount       uint    `json:"conn_quantity"`
			AverageSpeed    float32 `json:"average_speed"`
			AverageLatency  uint    `json:"average_latency"`
			AverageLossRate float32 `json:"average_packet_loss_rate"`
		}{
			PeriodID:  record1.PeriodID,
			ConnCount: record1.ConnCount,
			// TODO: 确定流量与延时的单位
			AverageSpeed:    float32(record1.TotalFlow * 1.0 / record1.AveLatency),
			AverageLatency:  record1.AveLatency,
			AverageLossRate: record1.LossRate,
		})
	}
	for _, record2 := range lastRecords {
		entity.Status = append(entity.Status, struct {
			PeriodID        uint    `json:"time_id"`
			ConnCount       uint    `json:"conn_quantity"`
			AverageSpeed    float32 `json:"average_speed"`
			AverageLatency  uint    `json:"average_latency"`
			AverageLossRate float32 `json:"average_packet_loss_rate"`
		}{
			PeriodID:  record2.PeriodID,
			ConnCount: record2.ConnCount,
			// TODO: 确定流量与延时的单位
			AverageSpeed:    float32(record2.TotalFlow * 1.0 / record2.AveLatency),
			AverageLatency:  record2.AveLatency,
			AverageLossRate: record2.LossRate,
		})
	}
	return entity, nil
}
