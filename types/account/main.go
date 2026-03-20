package accountTypes

type Account struct {
	Id            string `json:"id"`
	CompanyId     string `json:"companyId"`
	LineManagerId string `json:"lineManagerId"`
	IsAdmin       int8   `json:"isAdmin"`
	IsActive      int8   `json:"isActive"`
	Name          string `json:"name"`
	Surname       string `json:"surname"`
	Email         string `json:"email"`
	PhoneNumber   string `json:"phoneNumber"`
	Password      string `json:"-"`
}

func (account Account) IsEmpty() bool {
	return account.Id == ""
}
