package SQLController

import (
	"UserPortrait/etc"
	"UserPortrait/service"
	"fmt"
	"gorm.io/gorm"
	"sync"
)

var mutex sync.Mutex

func (s *SqlController) InsertUniverse() (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if newuni, ok := <-etc.UniverseChannel; ok {
			err = tx.Create(&newuni).Error
			if err != nil {
				fmt.Println("Hooks' error:", err)
				return fmt.Errorf("UID%v: update universe failed:%v\n", newuni.UserID, err)
			}
			fmt.Println("UID", newuni.UserID, ":update universe success")
			return nil
		} else {
			fmt.Println("update universe failed:UniverseChannel error")
			return fmt.Errorf("update universe failed:UniverseChannel error")
		}
	})
}

func (s *SqlController) UpdateUniverse() (err error) {
	mutex.Lock()
	defer mutex.Unlock()
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if uni, ok := <-etc.UniverseChannel; ok {
			err = tx.Table("universe").Where("user_id = ?", uni.UserID).Updates(map[string]interface{}{
				"flow":    uni.Flow,
				"latency": uni.Latency,
				"count":   uni.Count,
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

func (s *SqlController) AfterCreateUniverse(ip string) (err error) {
	locinfo, err := service.GetLocation(ip)
	if err != nil {
		return fmt.Errorf("Get location failed %v\n", err)
	}
	result := s.DB.Table("universe").Where("ip = ? AND district = ?", ip, "").Updates(map[string]interface{}{
		"district":  locinfo.Data.District,
		"latitude":  locinfo.Data.Lat,
		"longitude": locinfo.Data.Lng,
	}).Error
	return result
}
