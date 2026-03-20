package accessTypes

type Access struct {
	AccountId string `json:"accountId"`
	Register  string `json:"register"`
}

func (access Access) IsEmpty() bool {
	return access.AccountId == ""
}
