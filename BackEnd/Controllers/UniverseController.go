package Controllers

import (
	"UserPortrait/etc"
	"UserPortrait/functions"
	"fmt"
	"gorm.io/gorm"
	"sync"
)

var uniMutex sync.Mutex

// 插入新universe记录
func (s *SqlController) InsertUniverse(TableName string) (err error) {
	uniMutex.Lock()
	defer uniMutex.Unlock()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if newuni, ok := <-etc.UniverseChannel; ok {
			// 首先更新位置信息
			newuni.District, newuni.City, newuni.Longitude, newuni.Latitude, err = s.TransferLocationInfo(newuni.Ip)
			if err != nil {
				return fmt.Errorf("UID%v: update universe failed:%v\n", newuni.UserID, err)
			}
			// 创建完全体universe
			err = tx.Table(TableName).Create(&newuni).Error
			if err != nil {
				fmt.Println("Hooks' error:", err)
				return fmt.Errorf("UID%v: update universe failed:%v\n", newuni.UserID, err)
			}
			fmt.Println("UID", newuni.UserID, ":update universe success")
			return err
		} else {
			fmt.Println("update universe failed:UniverseChannel error")
			return fmt.Errorf("update universe failed:UniverseChannel error")
		}
	})
}

// 满足时空相同条件的universe更新
func (s *SqlController) UpdateUniverse(TableName string) (err error) {
	uniMutex.Lock()
	defer uniMutex.Unlock()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if uni, ok := <-etc.UniverseChannel; ok {
			err = tx.Table(TableName).Where("period_id = ? AND ip = ? AND date = ?", uni.PeriodID, uni.Ip, uni.Date).Updates(map[string]interface{}{
				"flow":      uni.Flow,
				"latency":   uni.Latency,
				"count":     uni.Count,
				"err_count": uni.ErrCount,
			}).Error
			if err != nil {
				fmt.Printf("UID%v Hooks' error:%v\n", uni.UserID, err)
				return fmt.Errorf("UID%v: update universe failed:%v\n", uni.UserID, err)
			}
			fmt.Println("UID", uni.UserID, ":update universe success")
			return nil
		} else {
			fmt.Println("update universe failed:UniverseChannel error")
			return fmt.Errorf("update universe failed:UniverseChannel error")
		}
	})
}

// 根据IP更新位置信息
func (s *SqlController) TransferLocationInfo(ip string) (string, string, float32, float32, error) {
	locinfo, err := GetLocation(ip)
	if err != nil {
		return "", "", 0, 0, fmt.Errorf("Get location failed %v\n", err)
	}
	roundLat := functions.RoundToFloat32(locinfo.Result.Location.Lat, 4)
	roundLng := functions.RoundToFloat32(locinfo.Result.Location.Lng, 4)
	return locinfo.Result.AdInfo.District, locinfo.Result.AdInfo.City, roundLat, roundLng, nil
}
