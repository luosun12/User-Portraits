package service

import (
	"UserPortrait/Controllers"
	"UserPortrait/functions"
	"UserPortrait/service/database"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func GetBaseStationInfo(c *gin.Context) {
	//TODO: 查询返回近24小时基站信息
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库连接失败,请重试",
		})
		fmt.Printf("login err:%v", err)
		return
	}
	stationId, _ := strconv.ParseUint(c.Query("station_id"), 10, 32)
	//fmt.Printf("Query Num Type:%T,%T", stationId, c.Query("station_id"))
	TableName := functions.ChooseTable(uint(stationId), "base_station")
	sql := Controllers.SqlController{DB: db}
	Yesterday, Today, lastPeriodId, currPeriodId, err := functions.GetDailyInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取时间信息失败,请重试",
		})
		return
	}
	result, err := sql.DailyStationRecords(uint(stationId), TableName, Yesterday, Today, lastPeriodId, currPeriodId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取基站信息失败,请重试",
		})
		fmt.Println("Get BaseStationInfo error:", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "获取基站信息成功",
		"data":    result,
	})
	return
}
