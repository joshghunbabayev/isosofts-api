package accountModels

import (
	"fmt"
	"isosofts-api/database"
	"isosofts-api/modules"
	accountTypes "isosofts-api/types/account"
	"strings"
)

type AccountModel struct {
}

func (*AccountModel) GenerateUniqueId() string {
	Id := modules.GenerateRandomString(30)

	var accountModel AccountModel

	br, _ := accountModel.GetById(Id)

	if br.IsEmpty() {
		return Id
	} else {
		return accountModel.GenerateUniqueId()
	}
}

func (*AccountModel) GetById(Id string) (accountTypes.Account, error) {
	db := database.GetDatabase()
	row := db.QueryRow(`
			SELECT * 
			FROM accounts
			WHERE id = ?
		`,
		Id,
	)

	var account accountTypes.Account

	err := row.Scan(
		&account.Id,
		&account.CompanyId,
		&account.LineManagerId,
		&account.IsAdmin,
		&account.IsActive,
		&account.Name,
		&account.Surname,
		&account.Email,
		&account.PhoneNumber,
		&account.Password,
	)

	return account, err
}

func (*AccountModel) GetAll(filters map[string]interface{}) ([]accountTypes.Account, error) {
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

	query := fmt.Sprintf(`
			SELECT * FROM accounts %s
		`,
		whereClause,
	)
	rows, err := db.Query(query, values...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []accountTypes.Account

	for rows.Next() {
		var account accountTypes.Account

		rows.Scan(
			&account.Id,
			&account.CompanyId,
			&account.LineManagerId,
			&account.IsAdmin,
			&account.IsActive,
			&account.Name,
			&account.Surname,
			&account.Email,
			&account.PhoneNumber,
			&account.Password,
		)
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (*AccountModel) Create(account accountTypes.Account) error {
	db := database.GetDatabase()
	_, err := db.Exec(`
			INSERT INTO accounts ( 
				"id",
				"companyId",
				"lineManagerId",
				"isAdmin",
				"isActive",
				"name",
				"surname",
				"email",
				"phoneNumber",
				"password"
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`,
		account.Id,
		account.CompanyId,
		account.LineManagerId,
		account.IsAdmin,
		account.IsActive,
		account.Name,
		account.Surname,
		account.Email,
		account.PhoneNumber,
		account.Password,
	)

	if err != nil {
		return err
	}

	return nil
}

func (*AccountModel) Update(Id string, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return nil
	}

	setClause := ""
	values := []interface{}{}

	for key, val := range fields {
		setClause += fmt.Sprintf(` "%s" = ?,`, key)
		values = append(values, val)
	}

	setClause = strings.TrimSuffix(setClause, ",")
	query := fmt.Sprintf(`
			UPDATE accounts 
			SET %s 
			WHERE "id" = ?
		`,
		setClause,
	)
	values = append(values, Id)

	db := database.GetDatabase()
	_, err := db.Exec(query, values...)
	return err
}
