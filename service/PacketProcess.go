package service

import (
	"UserPortrait/Controllers"
	"UserPortrait/etc"
	"UserPortrait/service/database"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strconv"
)

// 根据解包脚本更新universe信息，保证流时、空时分布的核心

func Packet2Universe(MAC string, IP string, datetime string, flow uint, latency uint) (err error) {
	db, err := database.InitDB()
	if err != nil {
		return err
	}
	var sql = Controllers.SqlController{DB: db}
	if MAC == "" {
		fmt.Println("MAC is empty")
		return fmt.Errorf("err:MAC is empty")
	}
	var NotExist bool
	var user etc.Userinfo
	var newuni etc.Universe
	var uni etc.Universe
	var ID uint = 0
	// 若已存在用户user保存了该MAC
	FoundUser := db.Table("user_info").Where("mac_info = ?", MAC).Take(&user)
	if FoundUser.Error != nil {
		NotExist = errors.Is(FoundUser.Error, gorm.ErrRecordNotFound)
		if !NotExist {
			fmt.Println(FoundUser.Error)
			return FoundUser.Error
		}
	}
	if NotExist {
		// 若该MAC未注册，创建用户NewUser，再更新或创建universe记录
		newUser := etc.Userinfo{MacInfo: MAC}
		db.Table("user_info").Create(&newUser)
		ID = newUser.ID
	} else {
		ID = user.ID
	}
	// 分解时段信息
	date, periodID, err := GetPeriod(datetime)
	if err != nil {
		return err
	}

	FoundUniverse := db.Table("universe").Where("user_id =? AND ip =? AND date =? AND period_id =?", ID, IP, date, periodID).Take(&uni)
	if FoundUniverse.Error != nil {
		if errors.Is(FoundUniverse.Error, gorm.ErrRecordNotFound) {
			newuni = etc.Universe{UserID: ID, Ip: IP, Date: date, Flow: flow, Latency: latency, PeriodID: periodID}
			etc.UniverseChannel <- newuni
			// 若该记录不存在，则创建记录
			if err = sql.InsertUniverse(); err != nil {
				return err
			}
		} else {
			return FoundUniverse.Error
		}
	} else {
		// 若该记录存在，即时间段与IP均重复，则累加flow，取latency平均，而后更新
		flow += uni.Flow
		latency = (uni.Latency + latency) / 2
		newuni = etc.Universe{UserID: uni.UserID, Ip: IP, Date: date, Flow: flow, Latency: latency, PeriodID: periodID, Count: uni.Count + 1}
		etc.UniverseChannel <- newuni
		if err := sql.UpdateUniverse(); err != nil {
			return err
		}
	}
	return nil
}

//获取日期和时段编码

func GetPeriod(t string) (string, uint, error) {

	if t == "" {
		return "", 0, fmt.Errorf("time is empty")
	}
	var hour = t[11:13]
	var periodId, _ = strconv.ParseUint(hour, 10, 64)
	if periodId >= 24 || periodId < 0 {
		return "", 0, fmt.Errorf("time is out of range")
	}
	return t[0:10], uint(periodId) + 1, nil
}
