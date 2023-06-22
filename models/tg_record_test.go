package models

import "testing"

func TestFindByLocationAndKeyword(t *testing.T) {
	locations := []string{"上海", "深圳"}
	keywords := []string{"个人", "兼职"}
	query := ""
	args := make([]interface{}, len(locations)*len(keywords)*2)
	for i, location := range locations {
		for j, keyword := range keywords {
			if i > 0 || j > 0 {
				query += " OR "
			}
			query += "(location LIKE ? AND keywords LIKE ?)"
			args[(i*len(keywords)+j)*2] = "%" + location + "%"
			args[(i*len(keywords)+j)*2+1] = "%" + keyword + "%"
		}
	}

	t.Log(query, args)
}
