package role

import (
	"testing"
)

func TestLoadRoles2DB(t *testing.T) {
	err := LoadRoles2DB()
	if err != nil {
		t.Error(err.Error())
	}
}
