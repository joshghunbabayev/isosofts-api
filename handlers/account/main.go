package accountHandlers

import (
	"fmt"
	accessModels "isosofts-api/models/access"
	accountModels "isosofts-api/models/account"
	companyModels "isosofts-api/models/company"
	accessTypes "isosofts-api/types/access"
	accountTypes "isosofts-api/types/account"
	companyTypes "isosofts-api/types/company"
	"net/http"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AccountHandlers struct {
}

func (*AccountHandlers) SignUp(c *gin.Context) {
	var body struct {
		Name            string `json:"name"`
		Surname         string `json:"surname"`
		Email           string `json:"email"`
		PhoneNumber     string `json:"phoneNumber"`
		CompanyName     string `json:"companyName"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	errs := make(map[string]string)

	if strings.TrimSpace(body.Name) == "" {
		errs["name"] = "Name is required"
	}

	if strings.TrimSpace(body.Surname) == "" {
		errs["surname"] = "Surname is required"
	}

	domain := ""
	if body.Email == "" {
		errs["email"] = "Email is required"
	} else {
		_, err := mail.ParseAddress(body.Email)
		if err != nil {
			errs["email"] = "Invalid email format"
		} else {
			parts := strings.Split(body.Email, "@")
			if len(parts) > 1 {
				domain = strings.ToLower(parts[1])
			}

			var companyModel companyModels.CompanyModel
			existingCompanies, _ := companyModel.GetAll(map[string]interface{}{
				"domain": domain,
			})

			if len(existingCompanies) > 0 {
				errs["email"] = "This domain is already registered"
			}
		}
	}

	if strings.TrimSpace(body.PhoneNumber) == "" {
		errs["phoneNumber"] = "Phone Number is required"
	}

	if strings.TrimSpace(body.CompanyName) == "" {
		errs["companyName"] = "Company name is required"
	}

	if len(body.Password) < 6 {
		errs["password"] = "Password must be at least 6 characters long"
	}

	if body.ConfirmPassword != body.Password {
		errs["confirmPassword"] = "Passwords do not match"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	var companyModel companyModels.CompanyModel

	companyId := companyModel.GenerateUniqueId()

	companyModel.Create(companyTypes.Company{
		Id:       companyId,
		Name:     body.CompanyName,
		Domain:   domain,
		IsActive: 1,
	})

	var accountModel accountModels.AccountModel
	var accessModel accessModels.AccessModel

	newAccountId := accountModel.GenerateUniqueId()

	accountModel.Create(accountTypes.Account{
		Id:            newAccountId,
		CompanyId:     companyId,
		LineManagerId: "",
		IsAdmin:       1,
		IsActive:      1,
		Name:          body.Name,
		Surname:       body.Surname,
		Email:         body.Email,
		PhoneNumber:   body.PhoneNumber,
		Password:      body.Password,
	})

	registers := []string{
		"kpi", "br", "hsr", "leg", "eai", "ei", "tra", "doc",
		"ven", "cus", "fb", "ea", "moc", "fin", "aop", "mrm",
	}

	for _, reg := range registers {
		err := accessModel.Create(accessTypes.Access{
			AccountId: newAccountId,
			Register:  reg,
		})

		if err != nil {
			fmt.Printf("Error creating access for %s: %v\n", reg, err)
		}
	}

	algebraUrl := os.Getenv("ALGEBRA_API_URL") + "/api/isosofts/kpi/duplicate-defaults?companyId=" + companyId
	http.Get(algebraUrl)

	c.IndentedJSON(200, gin.H{"message": "Registration successful"})
}

func (*AccountHandlers) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	errs := make(map[string]string)

	if strings.TrimSpace(body.Email) == "" {
		errs["email"] = "Email is required"
	}
	if strings.TrimSpace(body.Password) == "" {
		errs["password"] = "Password is required"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	var accountModel accountModels.AccountModel

	accounts, err := accountModel.GetAll(map[string]interface{}{
		"email": body.Email,
	})

	if err != nil || len(accounts) == 0 {
		c.IndentedJSON(400, gin.H{"error": "Invalid email or password"})
		return
	}

	account := accounts[0]

	fmt.Print(account.Id)

	fmt.Print(account)
	if account.IsActive == 0 {
		c.IndentedJSON(403, gin.H{"error": "This account is inactive. Please contact your admin."})
		return
	}

	if account.Password != body.Password {
		c.IndentedJSON(400, gin.H{"error": "Invalid email or password"})
		return
	}

	jwtSecret := os.Getenv("JWT_SECRET")

	claims := jwt.MapClaims{
		"sub": account.Id,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		c.IndentedJSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	c.IndentedJSON(200, gin.H{
		"message": "Login successful",
		"token":   tokenString,
	})
}

func (*AccountHandlers) GetSelf(c *gin.Context) {
	account := c.MustGet("account").(accountTypes.Account)

	c.IndentedJSON(200, account)
}

func (*AccountHandlers) UpdatePasswordSelf(c *gin.Context) {
	account := c.MustGet("account").(accountTypes.Account)

	var body struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	errs := make(map[string]string)

	if body.CurrentPassword != account.Password {
		errs["currentPassword"] = "Current password is incorrect"
	}

	if len(body.NewPassword) < 6 {
		errs["newPassword"] = "New password must be at least 6 characters long"
	}

	if body.NewPassword != body.ConfirmPassword {
		errs["confirmPassword"] = "New passwords do not match"
	}

	if body.NewPassword == body.CurrentPassword && body.NewPassword != "" {
		errs["newPassword"] = "New password cannot be the same as current password"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	var accountModel accountModels.AccountModel
	err := accountModel.Update(account.Id, map[string]interface{}{
		"password": body.NewPassword,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(200, gin.H{"message": "Password updated successfully"})
}

func (*AccountHandlers) GetAll(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	isActiveFilter := c.Query("isActive")

	filters := map[string]interface{}{
		"companyId": admin.CompanyId,
	}

	if isActiveFilter != "" {
		filters["isActive"] = isActiveFilter
	}

	var accountModel accountModels.AccountModel
	accounts, err := accountModel.GetAll(filters)

	if err != nil {
		c.IndentedJSON(500, gin.H{"error": "Database error while retrieving accounts"})
		return
	}

	c.IndentedJSON(200, accounts)
}

func (*AccountHandlers) Create(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)

	var companyModel companyModels.CompanyModel
	company, err := companyModel.GetById(admin.CompanyId)
	if err != nil || company.IsEmpty() {
		c.JSON(500, gin.H{"error": "Admin company information not found"})
		return
	}

	var body struct {
		Name            string `json:"name"`
		Surname         string `json:"surname"`
		Email           string `json:"email"`
		PhoneNumber     string `json:"phoneNumber"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	errs := make(map[string]string)

	if strings.TrimSpace(body.Name) == "" {
		errs["name"] = "Name is required"
	}
	if strings.TrimSpace(body.Surname) == "" {
		errs["surname"] = "Surname is required"
	}

	if body.Email == "" {
		errs["email"] = "Email is required"
	} else {
		_, err := mail.ParseAddress(body.Email)
		if err != nil {
			errs["email"] = "Invalid email format"
		} else {
			// parts := strings.Split(body.Email, "@")
			// domain := strings.ToLower(parts[1])

			// if domain != strings.ToLower(company.Domain) {
			// errs["email"] = fmt.Sprintf("Email must belong to the company domain: %s", company.Domain)
			// }

			var accountModel accountModels.AccountModel
			existing, _ := accountModel.GetAll(map[string]interface{}{"email": body.Email})
			if len(existing) > 0 {
				errs["email"] = "This email is already registered"
			}
		}
	}

	if strings.TrimSpace(body.PhoneNumber) == "" {
		errs["phoneNumber"] = "Phone Number is required"
	}

	if len(body.Password) < 6 {
		errs["password"] = "Password must be at least 6 characters long"
	}

	if body.ConfirmPassword != body.Password {
		errs["confirmPassword"] = "Passwords do not match"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	var accountModel accountModels.AccountModel

	accountModel.Create(accountTypes.Account{
		Id:          accountModel.GenerateUniqueId(),
		CompanyId:   admin.CompanyId,
		IsAdmin:     0,
		IsActive:    1,
		Name:        body.Name,
		Surname:     body.Surname,
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		Password:    body.Password,
	})

	c.IndentedJSON(200, gin.H{"message": "Account created successfully"})
}

func (*AccountHandlers) GetOne(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")

	var accountModel accountModels.AccountModel
	account, err := accountModel.GetById(id)

	if err != nil || account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found in your company"})
		return
	}

	c.JSON(200, account)
}

func (*AccountHandlers) Update(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")

	var accountModel accountModels.AccountModel
	account, err := accountModel.GetById(id)
	if err != nil || account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found"})
		return
	}

	var body struct {
		Name        string `json:"name"`
		Surname     string `json:"surname"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	errs := make(map[string]string)

	if strings.TrimSpace(body.Name) == "" {
		errs["name"] = "Name is required"
	}
	if strings.TrimSpace(body.Surname) == "" {
		errs["surname"] = "Surname is required"
	}
	if strings.TrimSpace(body.PhoneNumber) == "" {
		errs["phoneNumber"] = "Phone Number is required"
	}

	if strings.ToLower(body.Email) != strings.ToLower(account.Email) {
		_, err := mail.ParseAddress(body.Email)
		if err != nil {
			errs["email"] = "Invalid email format"
		} else {
			var companyModel companyModels.CompanyModel
			company, _ := companyModel.GetById(admin.CompanyId)
			parts := strings.Split(body.Email, "@")
			domain := strings.ToLower(parts[1])

			if domain != strings.ToLower(company.Domain) {
				errs["email"] = fmt.Sprintf("Email must belong to the company domain: %s", company.Domain)
			}

			existing, _ := accountModel.GetAll(map[string]interface{}{"email": body.Email})
			if len(existing) > 0 {
				errs["email"] = "This email is already taken by another user"
			}
		}
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	err = accountModel.Update(id, map[string]interface{}{
		"name":        body.Name,
		"surname":     body.Surname,
		"email":       body.Email,
		"phoneNumber": body.PhoneNumber,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update staff member"})
		return
	}

	c.JSON(200, gin.H{"message": "Staff member updated successfully"})
}

func (*AccountHandlers) UpdateLineManager(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")

	var body struct {
		LineManagerId string `json:"lineManagerId"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	var accountModel accountModels.AccountModel

	account, err := accountModel.GetById(id)
	if err != nil || account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found"})
		return
	}

	if id == body.LineManagerId {
		c.JSON(400, gin.H{"error": "A user cannot be their own line manager"})
		return
	}

	if body.LineManagerId != "" {
		manager, err := accountModel.GetById(body.LineManagerId)
		if err != nil || manager.IsEmpty() || manager.CompanyId != admin.CompanyId {
			c.JSON(400, gin.H{"error": "Selected line manager is invalid or belongs to another company"})
			return
		}
	}

	err = accountModel.Update(id, map[string]interface{}{
		"lineManagerId": body.LineManagerId,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update line manager"})
		return
	}

	c.JSON(200, gin.H{"message": "Line manager updated successfully"})
}

func (*AccountHandlers) UpdatePassword(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")

	var accountModel accountModels.AccountModel
	account, err := accountModel.GetById(id)
	if err != nil || account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found"})
		return
	}

	var body struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	errs := make(map[string]string)

	if len(body.Password) < 6 {
		errs["password"] = "Password must be at least 6 characters long"
	}

	if body.Password != body.ConfirmPassword {
		errs["confirmPassword"] = "Passwords do not match"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	err = accountModel.Update(id, map[string]interface{}{
		"password": body.Password,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(200, gin.H{"message": "Staff member's password updated successfully"})
}

func (*AccountHandlers) Unactive(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")

	if admin.Id == id {
		c.JSON(400, gin.H{"error": "You cannot deactivate your own account"})
		return
	}

	var accountModel accountModels.AccountModel
	account, _ := accountModel.GetById(id)

	if account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found"})
		return
	}

	if err := accountModel.Update(id, map[string]interface{}{"isActive": 0}); err != nil {
		c.JSON(500, gin.H{"error": "Operation failed"})
		return
	}

	c.JSON(200, gin.H{"message": "Account deactivated"})
}

func (*AccountHandlers) Active(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")

	var accountModel accountModels.AccountModel
	account, _ := accountModel.GetById(id)

	if account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found"})
		return
	}

	if err := accountModel.Update(id, map[string]interface{}{"isActive": 1}); err != nil {
		c.JSON(500, gin.H{"error": "Operation failed"})
		return
	}

	c.JSON(200, gin.H{"message": "Account activated"})
}

func (*AccountHandlers) GetAccesses(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")

	var accountModel accountModels.AccountModel
	account, err := accountModel.GetById(id)
	if err != nil || account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found"})
		return
	}

	var accessModel accessModels.AccessModel
	accesses, err := accessModel.GetAll(map[string]interface{}{
		"accountId": id,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve accesses"})
		return
	}

	c.IndentedJSON(200, accesses)
}

func (*AccountHandlers) AddAccess(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")
	register := c.Query("register")

	if strings.TrimSpace(register) == "" {
		c.JSON(400, gin.H{"error": "Register query parameter is required"})
		return
	}

	var accountModel accountModels.AccountModel
	account, err := accountModel.GetById(id)
	if err != nil || account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found"})
		return
	}

	var accessModel accessModels.AccessModel
	err = accessModel.Create(accessTypes.Access{
		AccountId: id,
		Register:  register,
	})

	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to add access. It might already exist."})
		return
	}

	c.JSON(200, gin.H{"message": fmt.Sprintf("Access '%s' added successfully", register)})
}

func (*AccountHandlers) RemoveAccess(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)
	id := c.Param("id")
	register := c.Query("register")

	if strings.TrimSpace(register) == "" {
		c.JSON(400, gin.H{"error": "Register query parameter is required"})
		return
	}

	var accountModel accountModels.AccountModel
	account, _ := accountModel.GetById(id)
	if account.IsEmpty() || account.CompanyId != admin.CompanyId {
		c.JSON(404, gin.H{"error": "Staff member not found"})
		return
	}

	var accessModel accessModels.AccessModel
	err := accessModel.Delete(accessTypes.Access{
		AccountId: id,
		Register:  register,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to remove access"})
		return
	}

	c.JSON(200, gin.H{"message": "Access removed successfully"})
}
