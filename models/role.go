package models

import (
	"log"

	"gorm.io/gorm"
)

type Role struct {
	gorm.Model
	Name string `gorm:"uniqueIndex"`
	Desc string
}

func CreateRole(role *Role) error {
	if err := db.Create(role).Error; err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SearchRoleByName(name string) ([]*Role, error) {
	var roles []*Role
	if err := db.Where("name LIKE ?", "%"+name+"%").Find(&roles).Error; err != nil {
		log.Println(err)
		return nil, err
	}
	return roles, nil
}

func UpdateRole(role *Role) error {
	if err := db.Save(role).Error; err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func DeleteRole(id uint) error {
	if err := db.Delete(&Role{}, id).Error; err != nil {
		log.Println(err)
		return err
	}
	return nil
}
