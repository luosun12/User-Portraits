package service

import (
	"UserPortrait/Controllers"
	"UserPortrait/service/database"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"time"
)

// 用户提交评分
func SubmitScore(c *gin.Context) {
	userID, _ := strconv.ParseUint(c.PostForm("id"), 10, 32)
	score, _ := strconv.ParseFloat(c.PostForm("score"), 32)
	date := time.Now().Format(time.DateOnly)
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库连接失败,请重试",
		})
		fmt.Printf("login err:%v", err)
		return
	}
	sql := Controllers.SqlController{DB: db}
	err = sql.FindScoreRecord(uint(userID), date)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = sql.InsertScore(uint(userID), float32(score))
		if err != nil {
			fmt.Println("UID ", userID, ": InsertScore err:", err)

			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "评分提交失败,请重试",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "评分提交成功",
		})
	} else {
		err = sql.UpdateScore(uint(userID), float32(score))
		if err != nil {
			fmt.Println("UID ", userID, ": UpdateScore err:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "评分提交失败,请重试",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "评分更新成功",
		})
	}
}

// 获取每日平均分的map列表
func GetAverageScore(c *gin.Context) {
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库连接失败,请重试",
		})
		fmt.Printf("login err:%v", err)
		return
	}
	sql := Controllers.SqlController{DB: db}
	scores, err := sql.AverageScoreByDate()
	if err != nil {
		fmt.Println("GetAverageScore err:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "获取平均分失败,请重试",
		})
		return
	}

	fmt.Println(scores)
	c.JSON(http.StatusOK, gin.H{
		"message": "获取平均分成功",
		"scores":  scores,
	})
}
