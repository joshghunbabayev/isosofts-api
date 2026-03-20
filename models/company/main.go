package companyModels

import (
	"fmt"
	"isosofts-api/database"
	"isosofts-api/modules"
	companyTypes "isosofts-api/types/company"
	"strings"
)

type CompanyModel struct {
}

func (*CompanyModel) GenerateUniqueId() string {
	Id := modules.GenerateRandomString(30)
	var companyModel CompanyModel
	br, _ := companyModel.GetById(Id)

	if br.IsEmpty() {
		return Id
	} else {
		return companyModel.GenerateUniqueId()
	}
}

func (*CompanyModel) GetById(Id string) (companyTypes.Company, error) {
	db := database.GetDatabase()
	row := db.QueryRow(`
			SELECT * FROM companies
			WHERE id = ?
		`,
		Id,
	)

	var company companyTypes.Company
	err := row.Scan(
		&company.Id,
		&company.Name,
		&company.Domain,
		&company.IsActive,
	)

	return company, err
}

func (*CompanyModel) GetAll(filters map[string]interface{}) ([]companyTypes.Company, error) {
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

	query := fmt.Sprintf(`SELECT * FROM companies %s`, whereClause)
	rows, err := db.Query(query, values...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var companies []companyTypes.Company
	for rows.Next() {
		var company companyTypes.Company
		rows.Scan(
			&company.Id,
			&company.Name,
			&company.Domain,
			&company.IsActive,
		)
		companies = append(companies, company)
	}

	return companies, nil
}

func (*CompanyModel) Create(company companyTypes.Company) error {
	db := database.GetDatabase()
	_, err := db.Exec(`
			INSERT INTO companies ( 
				"id",
				"name",
				"domain",
				"isActive"
			) VALUES (?, ?, ?, ?)
		`,
		company.Id,
		company.Name,
		company.Domain,
		company.IsActive,
	)

	if err != nil {
		return err
	}

	return nil
}

func (*CompanyModel) Update(Id string, fields map[string]interface{}) error {
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
	query := fmt.Sprintf(`UPDATE companies SET %s WHERE "id" = ?`, setClause)
	values = append(values, Id)

	db := database.GetDatabase()
	_, err := db.Exec(query, values...)
	return err
}
