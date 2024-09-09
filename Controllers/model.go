package Controllers

import "gorm.io/gorm"

type SqlController struct {
	DB *gorm.DB
}
