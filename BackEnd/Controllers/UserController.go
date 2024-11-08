package Controllers

import (
	"UserPortrait/etc"
)

func (s *SqlController) FindUserByMAC(mac string) (etc.Userinfo, error) {
	var user etc.Userinfo
	err := s.DB.Table("user_info").Where("mac_info = ?", mac).Take(&user).Error
	return user, err
}

func (s *SqlController) FindUserByName(name string) (etc.Userinfo, error) {
	var user etc.Userinfo
	err := s.DB.Table("user_info").Where("username = ?", name).Take(&user).Error
	return user, err
}

func (s *SqlController) InsertUser(user etc.Userinfo) {
	err := s.DB.Table("user_info").Create(&user).Error
	if err != nil {
		panic(err)
	}
}

func (s *SqlController) UpdateUserByID(id uint, name string, pswd string) {

	err := s.DB.Table("user_info").Where("id = ?", id).Updates(map[string]interface{}{
		"username": name, "password": pswd}).Error
	if err != nil {
		panic(err)
	}
}

// UserDailyFlow 用户：获取近24小时流量数据
func (s *SqlController) UserDailyFlow(userId uint, yesterday string, today string, lastID uint, currID uint) (etc.TrafficData, error) {
	var lastRecords []etc.Universe
	var currRecords []etc.Universe
	// 实例初始化
	var entity etc.TrafficData

	// 数据库条件遍历，获取昨日、今日近24小时记录
	Contents1 := s.DB.Model(&[]etc.Universe{}).Table("universe1").Select("user_id,date,period_id,SUM(flow) as flow")
	Contents2 := s.DB.Model(&[]etc.Universe{}).Table("universe1").Select("user_id,date,period_id,SUM(flow) as flow")
	err1 := Contents1.Where("user_id =? AND date =? AND period_id >=? AND period_id <=?", userId, yesterday, int(lastID), 24).Group("user_id, date, period_id").Order("period_id").Find(&lastRecords).Error
	if err1 != nil {
		return entity, err1
	}
	err2 := Contents2.Where("user_id =? AND date =? AND period_id >=? AND period_id <=?", userId, today, 1, int(currID)).Group("user_id, date, period_id").Order("period_id").Find(&currRecords).Error
	if err2 != nil {
		return entity, err2
	}

	for _, record1 := range currRecords {
		entity.Traffic[record1.PeriodID-1] = record1.Flow
	}
	for _, record2 := range lastRecords {
		entity.Traffic[record2.PeriodID-1] = record2.Flow
	}
	return entity, nil
}

// UserFreqLoc 统计用户常去地点
func (s *SqlController) UserFreqLoc(userId uint, tableName string) ([]etc.FreqLocation, error) {
	var entity []etc.FreqLocation
	err := s.DB.Model(&[]etc.Universe{}).Table(tableName).Select("city as name, COUNT(city) as count,latitude as lat,longitude as lng").Where("user_id = ?", userId).Group("city,latitude,longitude").Order("city").Find(&entity).Error
	return entity, err
}
