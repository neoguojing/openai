package role

import (
	"os"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestConvert(t *testing.T) {

	Convert("./role.txt", "./role.yaml")

	yamlFile, err := os.Open("./role.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer yamlFile.Close()

	var roles Roles
	yamlDecoder := yaml.NewDecoder(yamlFile)
	err = yamlDecoder.Decode(&roles)
	if err != nil {
		t.Fatal(err)
	}

	if roles["充当 Linux 终端"] == "" {
		t.Errorf("Expected Role1 description to be no empty, but got empty")
	}

	if roles["充当 JavaScript 控制台"] == "" {
		t.Errorf("Expected Role2 description to be no empty, but got empty")
	}

	if roles["充当“电影/书籍/任何东西”中的“角色”"] == "" {
		t.Errorf("Expected Role3 description to be no empty, but got empty")
	}
}
