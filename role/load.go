package role

import (
	"os"

	"github.com/neoguojing/log"

	"github.com/neoguojing/gormboot/v2"
	"github.com/neoguojing/openai/models"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

type Roles map[string]string

func init() {
	db = gormboot.DefaultDB.AutoMigrate().DB()
}

func LoadRoles2DB() error {
	var roles Roles
	yamlFile, err := os.Open("./role.yaml")
	if err != nil {
		panic(err)
	}
	defer yamlFile.Close()
	yamlDecoder := yaml.NewDecoder(yamlFile)
	err = yamlDecoder.Decode(&roles)
	if err != nil {
		panic(err)
	}

	cnt, err := models.CountRoles()
	if err != nil {
		panic(err)
	}
	if int64(len(roles)) <= cnt {
		return nil
	}

	for role, desc := range roles {
		user := models.Role{Name: role, Desc: desc}
		if err := db.Create(&user).Error; err != nil {
			log.Error(err.Error())
			return err
		}
	}

	return nil
}
