package models

import (
	"github.com/neoguojing/log"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name string `gorm:"uniqueIndex"`
	Desc string
}

func CountRoles() (int64, error) {
	var count int64
	if err := db.Model(&Role{}).Count(&count).Error; err != nil {
		log.Error(err.Error())
		return 0, err
	}
	return count, nil
}

func CreateRole(role *Role) error {
	if err := db.Create(role).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func SearchRoleByName(name string) ([]*Role, error) {
	var roles []*Role
	if err := db.Where("name LIKE ?", "%"+name+"%").Find(&roles).Error; err != nil {
		log.Error(err.Error())
		return nil, err
	}
	return roles, nil
}

func UpdateRole(role *Role) error {
	if err := db.Save(role).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}

func DeleteRole(id uint) error {
	if err := db.Delete(&Role{}, id).Error; err != nil {
		log.Error(err.Error())
		return err
	}
	return nil
}
