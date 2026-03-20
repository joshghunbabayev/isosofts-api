package accessModels

import (
	"fmt"
	"isosofts-api/database"
	accessTypes "isosofts-api/types/access"
	"strings"
)

type AccessModel struct {
}

func (*AccessModel) GetAll(filters map[string]interface{}) ([]accessTypes.Access, error) {
	db := database.GetDatabase()
	whereClause := ""
	values := []interface{}{}

	if len(filters) > 0 {
		whereParts := []string{}
		for key, val := range filters {
			whereParts = append(whereParts, fmt.Sprintf(`"%s" = ?`, key))
			values = append(values, val)
		}
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	query := fmt.Sprintf(`SELECT accountId, register FROM accesses %s`, whereClause)
	rows, err := db.Query(query, values...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accesses []accessTypes.Access
	for rows.Next() {
		var access accessTypes.Access
		rows.Scan(
			&access.AccountId,
			&access.Register,
		)
		accesses = append(accesses, access)
	}

	return accesses, nil
}

func (*AccessModel) Create(access accessTypes.Access) error {
	db := database.GetDatabase()
	_, err := db.Exec(`
			INSERT INTO accesses (
				"accountId",
				"register"
			) VALUES (?, ?)
		`,
		access.AccountId,
		access.Register,
	)
	return err
}

func (*AccessModel) Delete(access accessTypes.Access) error {
	db := database.GetDatabase()
	_, err := db.Exec(`
			DELETE FROM accesses 
			WHERE accountId = ? AND register = ?
		`, access.AccountId, access.Register)
	return err
}
