package companyTypes

type Company struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Domain   string `json:"domain"`
	IsActive int8   `json:"isActive"`
}

func (company Company) IsEmpty() bool {
	return company.Id == ""
}
