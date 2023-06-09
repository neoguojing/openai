package models

import "testing"

func TestSearchRoleByName(t *testing.T) {
	roles, err := SearchRoleByName("职业")
	if err != nil {
		t.Error(err.Error())
	}

	if len(roles) == 0 {
		t.Errorf("should not be zero record")
	}

	t.Log(roles[0].Name)
	t.Log(roles[0].Desc)
}
