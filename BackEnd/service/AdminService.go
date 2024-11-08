package service

import (
	"UserPortrait/Controllers"
	"UserPortrait/etc"
	"UserPortrait/service/database"
	"UserPortrait/token"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
)

func AdminRegister(c *gin.Context) {
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库连接失败,请重试",
		})
		fmt.Printf("register err:%v\n", err)
		return
	}
	var sql = Controllers.SqlController{DB: db}
	var administrator etc.Admininfo
	newName := c.PostForm("admin_name")
	newPswd := c.PostForm("password")
	pswd, _ := bcrypt.GenerateFromPassword([]byte(newPswd), bcrypt.DefaultCost)
	if newName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "用户名不能为空",
		})
		fmt.Printf("register err:Empty Admin Name\n")
		return
	}
	if newPswd == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "密码不能为空",
		})
		fmt.Printf("register err:Empty Password\n")
		return
	}

	// 无问题，更进结构体
	administrator.Adminname = newName
	administrator.Password = string(pswd)
	err = sql.InsertAdmin(administrator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "注册失败,请重试",
		})
		fmt.Printf("register err:%v\n", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "注册成功",
	})
	return
}

func AdminLogin(c *gin.Context) {
	db, err := database.InitDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库连接失败,请重试",
		})
		fmt.Printf("login err:%v\n", err)
		return
	}
	var sql = Controllers.SqlController{DB: db}
	var admin etc.Admininfo
	adminName := c.PostForm("admin_name")
	adminPswd := c.PostForm("password")
	if adminName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "用户名不能为空",
		})
		fmt.Printf("login err:Empty Admin Name\n")
		return
	}
	if adminPswd == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "密码不能为空",
		})
		fmt.Printf("login err:Empty Password\n")
		return
	}
	admin.Adminname = adminName
	admin.Password = adminPswd
	result, err := sql.FindAdminByName(admin.Adminname)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "用户名或密码错误",
			})
			fmt.Printf("login err:%v\n", err)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "服务器内部错误",
			})
			fmt.Printf("login err:%v\n", err)
		}
	}
	if bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(admin.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "用户名或密码错误",
		})
		fmt.Printf("login err:Password Error\n")
		return
	}
	geneToken, errt := token.GenerateAdminToken(result.ID)
	if errt != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "服务器内部错误",
		})
		fmt.Printf("login err:%v\n", errt)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "登录成功",
		"token":   geneToken,
	})
	return

}
