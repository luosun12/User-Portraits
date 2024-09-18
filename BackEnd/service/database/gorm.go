package database

import (
	"UserPortrait/configs"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Gorm会自动创建和管理连接池，因此不需手动关闭连接

func InitDB() (*gorm.DB, error) {
	dsn := configs.DBUser + ":" + configs.DBPassword + "@tcp(" + configs.DBHost + ")/test?charset=utf8mb4&parseTime=True"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		errors := fmt.Errorf("failed to connect database, %v", err)
		return nil, errors
	}
	return db, nil
}
