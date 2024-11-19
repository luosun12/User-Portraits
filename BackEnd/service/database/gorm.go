package database

import (
	"UserPortrait/configs"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Gorm会自动创建和管理连接池，因此不需手动关闭连接
var db *gorm.DB

func InitDB() (*gorm.DB, error) {
	if db != nil {
		return db, nil
	} else {
		dsn := configs.DBUser + ":" + configs.DBPassword + "@tcp(" + configs.DBHost + ")/IUPG?charset=utf8mb4&parseTime=True"
		newDb, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			DisableForeignKeyConstraintWhenMigrating: true,
			Logger:                                   logger.Default.LogMode(logger.Warn),
		})
		db = newDb
		if err != nil {
			errors := fmt.Errorf("failed to connect database, %v", err)
			return nil, errors
		}
		sqlDB, err := newDb.DB()
		if err != nil {
			errors := fmt.Errorf("failed to get database connection, %v", err)
			return nil, errors
		}
		sqlDB.SetMaxOpenConns(configs.DBMaxOpenConns)
		sqlDB.SetMaxIdleConns(configs.DBMaxIdleConns)
		sqlDB.SetConnMaxLifetime(configs.DBConnMaxLifetime)
		return newDb, nil
	}
}
