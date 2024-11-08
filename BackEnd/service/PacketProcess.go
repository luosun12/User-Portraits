package service

import (
	"UserPortrait/Controllers"
	"UserPortrait/etc"
	"UserPortrait/service/database"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

// 根据解包脚本更新universe信息，保证流时、空时分布的核心
// 需别处增加判定station_id的部分

func Packet2Universe(stationId uint, lossFlag bool, MAC string, IP string, datetime string, flow uint, latency uint) (err error) {
	universeTable := etc.ChooseTable(stationId, "universe")
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
	date, periodID, err := etc.GetPeriod(datetime)
	if err != nil {
		return err
	}

	FoundUniverse := db.Table(universeTable).Where("user_id =? AND ip =? AND date =? AND period_id =?", ID, IP, date, periodID).Take(&uni)
	if FoundUniverse.Error != nil {
		if errors.Is(FoundUniverse.Error, gorm.ErrRecordNotFound) {
			newuni = etc.Universe{UserID: ID, Ip: IP, Date: date, Flow: flow, Latency: latency, PeriodID: periodID}
			etc.UniverseChannel <- newuni
			// 若该记录不存在，则创建记录
			if err = sql.InsertUniverse(universeTable); err != nil {
				return err
			}
		} else {
			return FoundUniverse.Error
		}
	} else {
		// 若该记录存在，即时间段与IP均重复，则累加flow，取latency平均；若LossFlag为true，则err_count+1
		err_count := uni.ErrCount
		flow += uni.Flow
		if lossFlag {
			err_count += 1
			newuni = etc.Universe{UserID: uni.UserID, Ip: IP, Date: date, Flow: flow, Latency: uni.Latency, PeriodID: periodID, Count: uni.Count + 1, ErrCount: err_count}
		} else {
			latency = (uni.Latency*uni.Count + latency) / (uni.Count + 1)
			newuni = etc.Universe{UserID: uni.UserID, Ip: IP, Date: date, Flow: flow, Latency: latency, PeriodID: periodID, Count: uni.Count + 1, ErrCount: uni.ErrCount}
		}
		etc.UniverseChannel <- newuni
		if err := sql.UpdateUniverse(universeTable); err != nil {
			return err
		}

	}
	return nil
}

// 在Universe更改后执行，更新基站记录

func Packet2BaseStation(stationId uint, lossFlag bool, datetime string, flow uint, latency uint) error {
	stationTable := etc.ChooseTable(stationId, "base_station")
	db, err := database.InitDB()
	if err != nil {
		return err
	}
	var sql = Controllers.SqlController{DB: db}
	date, periodID, err := etc.GetPeriod(datetime)
	if err != nil {
		fmt.Println(err)
		return err
	}
	var newrecord = etc.BaseStation{Date: date, PeriodID: periodID, TotalFlow: flow, AveLatency: latency}
	if lossFlag {
		newrecord.ErrCount = 1
	}
	etc.StationChannel <- newrecord
	err = sql.UpdateStationAfterUni(stationTable)
	return err
}
