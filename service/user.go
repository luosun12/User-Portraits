package service

import (
	"UserPortrait/etc"
	"UserPortrait/service/SQLController"
	"UserPortrait/service/database"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
)

// 用户注册

func Register(c *gin.Context) {
	context := c
	db, err := database.InitDB()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库连接失败,请重试",
		})
		fmt.Printf("register err:%v\n", err)
		return
	}
	var sql = SQLController.SqlController{DB: db}
	var user etc.Userinfo

	newname := context.PostForm("username")
	newMAC := context.PostForm("MAC")
	if newname == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "用户名不能为空",
		})
		fmt.Printf("register err:bad newname\n")
		return
	}
	user, err = sql.FindUserByMAC(newMAC)
	if err != nil {
		// MAC不存在，可以注册
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("register: 用户 %v 不存在，可以注册\n", newname)
			pswd, _ := bcrypt.GenerateFromPassword([]byte(context.PostForm("password")), bcrypt.DefaultCost)
			user = etc.Userinfo{Username: newname, Password: string(pswd), MacInfo: newMAC}
			sql.InsertUser(user)
			context.JSON(http.StatusOK, gin.H{
				"message": "恭喜您，注册成功！",
			})
			return
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{
				"message": "数据库查询错误，请重试",
			})
			fmt.Printf("register err:%v\n", err)
			return
		}
	}
	if user.MacInfo != "" {
		// MAC,username 均存在，无需注册
		if user.Username != "" {
			context.JSON(http.StatusUnauthorized, gin.H{
				"message": "用户名已存在",
			})
			fmt.Printf("register err:user %v has existed\n", user.Username)
			return
		} else {
			// MAC存在，而无user信息，仍需注册，此时用Update替换空串
			sql.UpdateUserByID(user.ID, newname, context.PostForm("password"))
			context.JSON(http.StatusOK, gin.H{
				"message": "恭喜您，注册成功！",
			})
			fmt.Printf("register: 新用户%v注册成功\n", newname)
			return
		}
	}
}

// 用户登录
func Login(c *gin.Context) {
	context := c
	db, err := database.InitDB()
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"message": "数据库连接失败,请重试",
		})
		fmt.Printf("login err:%v", err)
		return
	}
	var sql = SQLController.SqlController{DB: db}
	var user etc.Userinfo
	username := context.PostForm("username")
	user, err = sql.FindUserByName(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			context.JSON(http.StatusUnauthorized, gin.H{
				"message": "用户名不存在,请注册！",
			})
			fmt.Printf("login err:user %v is not existed\n", username)
			return
		} else {
			context.JSON(http.StatusInternalServerError, gin.H{
				"message": "数据库查询错误，请重试",
			})
			fmt.Printf("login err:%v\n", err)
			return
		}
	} else {
		if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(context.PostForm("password"))) == nil {
			context.JSON(http.StatusOK, gin.H{
				"message": "登录成功",
			})
			fmt.Printf("login: user %v login success\n", username)
			return
		} else {
			context.JSON(http.StatusUnauthorized, gin.H{
				"message": "密码错误",
			})
			fmt.Printf("login err:user %v password is wrong\n", username)
			return
		}
	}
}

// 用户头像上传
func UploadAvatar(c *gin.Context) {
	context := c
	image, err := context.FormFile("avatar")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"message": "请选择图片文件",
		})
		return
	}
	imageType := strings.Split(image.Filename, ".")[1]
	if imageType == "jpg" || imageType == "jpeg" || imageType == "png" {
		userid := context.PostForm("user_id")
		newfilename := fmt.Sprintf("%v.%v", userid, imageType)
		dst := filepath.Join(etc.AvatarUploadPath, newfilename)
		if err := c.SaveUploadedFile(image, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "头像保存成功"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请上传jpg/jpeg/png格式图片"})
	}
}

// TODO: 用户基本信息获取
func GetBasicInfo(c *gin.Context) {
	context := c
	userid := context.Param("id")
	fmt.Println(userid)
	return
}

// 插入数据测试
func Ping(c *gin.Context) {
	cc := c
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		cc.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
		defer wg.Done()
	}()
}
