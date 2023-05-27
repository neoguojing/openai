package role

import (
	"log"
	"os"

	"github.com/neoguojing/gormboot/v2"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

type Roles map[string]string

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

func init() {
	factory := gormboot.New(gormboot.DefaultSqliteConfig("../sqlite3.db"))
	factory.RegisterModel(&Role{})
	db = factory.AutoMigrate().DB()
}

func LoadRoles2DB() error {
	var roles Roles
	yamlFile, err := os.Open("./role.yaml")
	if err != nil {
		log.Fatal(err)
	}
	defer yamlFile.Close()
	yamlDecoder := yaml.NewDecoder(yamlFile)
	err = yamlDecoder.Decode(&roles)
	if err != nil {
		log.Fatal(err)
	}

	for role, desc := range roles {
		user := Role{Name: role, Desc: desc}
		if err := db.Create(&user).Error; err != nil {
			log.Println(err)
			return err
		}
	}

	return nil
}
