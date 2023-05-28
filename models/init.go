package models

import (
	"github.com/neoguojing/gormboot/v2"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func init() {
	gormboot.DefaultDB.RegisterModel(&Role{}, &ChatRecord{})
	db = gormboot.DefaultDB.AutoMigrate().DB()
}
