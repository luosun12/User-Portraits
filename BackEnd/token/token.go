package token

import (
	"UserPortrait/configs"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"strconv"
	"strings"
	"time"
)

func GenerateUserToken(userId uint) (string, error) {
	tokenLifespan := configs.TOKEN_LIFESPAN
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["user_id"] = userId
	claims["salt"] = configs.Salt
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(tokenLifespan)).Unix()
	userToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return userToken.SignedString([]byte(configs.TOKEN_SECRET))
}

//TODO:管理员token需要接口

func GenerateAdminToken(adminId uint) (string, error) {
	tokenLifespan := configs.TOKEN_LIFESPAN
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["admin_id"] = adminId
	claims["admin_salt"] = configs.AdminSalt
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(tokenLifespan)).Unix()
	adminToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return adminToken.SignedString([]byte(configs.TOKEN_SECRET))
}

func TokenValid(c *gin.Context) error {
	tokenString := ExtractToken(c)
	if tokenString == "" {
		return fmt.Errorf("token is missing")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(configs.TOKEN_SECRET), nil
	})
	if err != nil {
		return err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["salt"] == configs.Salt && claims["authorized"] == true {
			fmt.Printf("UID %v\n:Token Valid\n", claims["user_id"])
			return nil
		}
	} else {
		return errors.New("invalid token")
	}
	exp, ok := token.Claims.(jwt.MapClaims)["exp"].(float64)
	if !ok {
		return errors.New("invalid exp claim")
	}
	if time.Now().Unix() > int64(exp) {
		return errors.New("token has expired")
	}
	return nil
}

// 从请求头中获取token
func ExtractToken(c *gin.Context) string {
	bearerToken := c.GetHeader("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		token := strings.Split(bearerToken, " ")[1]
		return token
	}
	return ""
}

// 从jwt中解析出user_id
func ExtractTokenID(c *gin.Context) (uint, error) {
	tokenString := ExtractToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(configs.TOKEN_SECRET), nil
	})
	if err != nil {
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	// 如果jwt有效，将user_id转换为浮点数字符串，然后再转换为 uint32
	if ok && token.Valid {
		uid, err := strconv.ParseUint(fmt.Sprintf("%.0f", claims["user_id"]), 10, 32)
		if err != nil {
			return 0, err
		}
		return uint(uid), nil
	}

	return 0, nil
}
