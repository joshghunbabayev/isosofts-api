package superAdminHandlers

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

type SuperAdminHandlers struct{}

func (*SuperAdminHandlers) Login(c *gin.Context) {
	var body struct {
		Password1 string `json:"password1"`
		Password2 string `json:"password2"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	if strings.TrimSpace(body.Password1) != os.Getenv("PASSWORD1") || strings.TrimSpace(body.Password2) != os.Getenv("PASSWORD2") {
		c.Status(401)
		return
	}

	adminId := os.Getenv("ADMIN_ID")
	jwtSecret := os.Getenv("JWT_SECRET_FOR_ADMIN")

	claims := jwt.MapClaims{
		"sub": adminId,
		"exp": time.Now().Add(time.Hour * 72).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(jwtSecret))

	c.JSON(200, gin.H{"token": tokenString})
}

func (*SuperAdminHandlers) GetAllCompanies(c *gin.Context) {
	isActiveFilter := c.Query("isActive")
	filters := make(map[string]interface{})

	if isActiveFilter != "" {
		filters["isActive"] = isActiveFilter
	}

	var companyModel companyModels.CompanyModel
	companies, err := companyModel.GetAll(filters)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}
	c.JSON(200, companies)
}

func (*SuperAdminHandlers) CreateCompany(c *gin.Context) {
	var body struct {
		Name   string `json:"name"`
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid format"})
		return
	}

	errs := make(map[string]string)

	if strings.TrimSpace(body.Name) == "" {
		errs["name"] = "Company name is required"
	}

	if strings.TrimSpace(body.Domain) == "" {
		errs["domain"] = "Domain is required"
	} else if !strings.Contains(body.Domain, ".") {
		errs["domain"] = "Invalid domain format (e.g., company.com)"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	var companyModel companyModels.CompanyModel

	companyId := companyModel.GenerateUniqueId()

	err := companyModel.Create(companyTypes.Company{
		Id:       companyId,
		Name:     body.Name,
		Domain:   strings.ToLower(body.Domain),
		IsActive: 1,
	})

	algebraUrl := os.Getenv("ALGEBRA_API_URL") + "/api/isosofts/kpi/duplicate-defaults?companyId=" + companyId
	http.Get(algebraUrl)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create company"})
		return
	}
	c.JSON(201, gin.H{"message": "Company created successfully"})
}

func (*SuperAdminHandlers) UpdateCompany(c *gin.Context) {
	id := c.Param("companyId")
	var companyModel companyModels.CompanyModel

	company, err := companyModel.GetById(id)
	if err != nil || company.IsEmpty() {
		c.JSON(404, gin.H{"error": "Company not found"})
		return
	}

	var body struct {
		Name   string `json:"name"`
		Domain string `json:"domain"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid format"})
		return
	}

	errs := make(map[string]string)
	if strings.TrimSpace(body.Name) == "" {
		errs["name"] = "Company name is required"
	}
	if strings.TrimSpace(body.Domain) == "" {
		errs["domain"] = "Domain is required"
	} else if !strings.Contains(body.Domain, ".") {
		errs["domain"] = "Invalid domain format"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	err = companyModel.Update(id, map[string]interface{}{
		"name":   body.Name,
		"domain": strings.ToLower(body.Domain),
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update company"})
		return
	}

	c.JSON(200, gin.H{"message": "Company updated successfully"})
}

func (*SuperAdminHandlers) ActivateCompany(c *gin.Context) {
	companyId := c.Param("companyId")
	var companyModel companyModels.CompanyModel
	if err := companyModel.Update(companyId, map[string]interface{}{"isActive": 1}); err != nil {
		c.Status(500)
		return
	}
	c.JSON(200, gin.H{"message": "Company activated"})
}

func (*SuperAdminHandlers) DeactivateCompany(c *gin.Context) {
	companyId := c.Param("companyId")
	var companyModel companyModels.CompanyModel
	if err := companyModel.Update(companyId, map[string]interface{}{"isActive": 0}); err != nil {
		c.Status(500)
		return
	}
	c.JSON(200, gin.H{"message": "Company deactivated"})
}

func (*SuperAdminHandlers) GetAllAccounts(c *gin.Context) {
	isActiveFilter := c.Query("isActive")
	companyIdFilter := c.Query("companyId")
	filters := make(map[string]interface{})

	if isActiveFilter != "" {
		filters["isActive"] = isActiveFilter
	}
	if companyIdFilter != "" {
		filters["companyId"] = companyIdFilter
	}

	var accountModel accountModels.AccountModel
	accounts, err := accountModel.GetAll(filters)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}
	c.JSON(200, accounts)
}

func (*SuperAdminHandlers) GetCompanyAccounts(c *gin.Context) {
	companyId := c.Param("companyId")
	isActiveFilter := c.Query("isActive")

	filters := map[string]interface{}{"companyId": companyId}
	if isActiveFilter != "" {
		filters["isActive"] = isActiveFilter
	}

	var accountModel accountModels.AccountModel
	accounts, err := accountModel.GetAll(filters)
	if err != nil {
		c.Status(500)
		return
	}
	c.JSON(200, accounts)
}

func (*SuperAdminHandlers) Create(c *gin.Context) {
	companyId := c.Param("companyId")
	var body struct {
		Name            string `json:"name"`
		Surname         string `json:"surname"`
		Email           string `json:"email"`
		PhoneNumber     string `json:"phoneNumber"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
		IsAdmin         int8   `json:"isAdmin"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid format"})
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

	var companyModel companyModels.CompanyModel
	company, _ := companyModel.GetById(companyId)

	if body.Email == "" {
		errs["email"] = "Email is required"
	} else {
		_, err := mail.ParseAddress(body.Email)
		if err != nil {
			errs["email"] = "Invalid email format"
		} else {
			parts := strings.Split(body.Email, "@")
			domain := strings.ToLower(parts[1])
			if domain != strings.ToLower(company.Domain) {
				errs["email"] = fmt.Sprintf("Email must belong to domain: %s", company.Domain)
			}
			var accountModel accountModels.AccountModel
			existing, _ := accountModel.GetAll(map[string]interface{}{"email": body.Email})
			if len(existing) > 0 {
				errs["email"] = "This email is already registered"
			}
		}
	}

	if len(body.Password) < 6 {
		errs["password"] = "Password must be at least 6 characters"
	}
	if body.Password != body.ConfirmPassword {
		errs["confirmPassword"] = "Passwords do not match"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	var accountModel accountModels.AccountModel
	err := accountModel.Create(accountTypes.Account{
		Id:          accountModel.GenerateUniqueId(),
		CompanyId:   companyId,
		IsAdmin:     body.IsAdmin,
		IsActive:    1,
		Name:        body.Name,
		Surname:     body.Surname,
		Email:       body.Email,
		PhoneNumber: body.PhoneNumber,
		Password:    body.Password,
	})

	if err != nil {
		c.Status(500)
		return
	}
	c.JSON(201, gin.H{"message": "Account created"})
}

func (*SuperAdminHandlers) GetOne(c *gin.Context) {
	id := c.Param("id")
	var accountModel accountModels.AccountModel
	account, err := accountModel.GetById(id)
	if err != nil || account.IsEmpty() {
		c.Status(404)
		return
	}
	c.JSON(200, account)
}

func (*SuperAdminHandlers) Update(c *gin.Context) {
	id := c.Param("id")
	var accountModel accountModels.AccountModel
	account, err := accountModel.GetById(id)
	if err != nil || account.IsEmpty() {
		c.Status(404)
		return
	}

	var body struct {
		Name        string `json:"name"`
		Surname     string `json:"surname"`
		Email       string `json:"email"`
		PhoneNumber string `json:"phoneNumber"`
		IsAdmin     int8   `json:"isAdmin"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.Status(400)
		return
	}

	errs := make(map[string]string)
	if strings.TrimSpace(body.Name) == "" {
		errs["name"] = "Name is required"
	}
	if strings.TrimSpace(body.Surname) == "" {
		errs["surname"] = "Surname is required"
	}

	if strings.ToLower(body.Email) != strings.ToLower(account.Email) {
		_, err := mail.ParseAddress(body.Email)
		if err != nil {
			errs["email"] = "Invalid email format"
		} else {
			var companyModel companyModels.CompanyModel
			company, _ := companyModel.GetById(account.CompanyId)
			parts := strings.Split(body.Email, "@")
			if strings.ToLower(parts[1]) != strings.ToLower(company.Domain) {
				errs["email"] = fmt.Sprintf("Email must belong to domain: %s", company.Domain)
			}
			existing, _ := accountModel.GetAll(map[string]interface{}{"email": body.Email})
			if len(existing) > 0 {
				errs["email"] = "Email already taken"
			}
		}
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	accountModel.Update(id, map[string]interface{}{
		"name":        body.Name,
		"surname":     body.Surname,
		"email":       body.Email,
		"phoneNumber": body.PhoneNumber,
		"isAdmin":     body.IsAdmin,
	})
	c.JSON(200, gin.H{"message": "Updated"})
}

func (*SuperAdminHandlers) UpdateLineManager(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		LineManagerId string `json:"lineManagerId"`
	}
	c.ShouldBindJSON(&body)

	var accountModel accountModels.AccountModel
	account, _ := accountModel.GetById(id)
	if account.IsEmpty() {
		c.Status(404)
		return
	}

	if body.LineManagerId != "" {
		manager, _ := accountModel.GetById(body.LineManagerId)
		if manager.IsEmpty() || manager.CompanyId != account.CompanyId {
			c.JSON(400, gin.H{"error": "Invalid line manager for this company"})
			return
		}
	}

	accountModel.Update(id, map[string]interface{}{"lineManagerId": body.LineManagerId})
	c.JSON(200, gin.H{"message": "Line Manager updated"})
}

func (*SuperAdminHandlers) UpdatePassword(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	c.ShouldBindJSON(&body)

	errs := make(map[string]string)
	if len(body.Password) < 6 {
		errs["password"] = "Minimum 6 characters required"
	}
	if body.Password != body.ConfirmPassword {
		errs["confirmPassword"] = "Passwords do not match"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	var accountModel accountModels.AccountModel
	accountModel.Update(id, map[string]interface{}{"password": body.Password})
	c.JSON(200, gin.H{"message": "Password updated"})
}

func (*SuperAdminHandlers) Active(c *gin.Context) {
	id := c.Param("id")
	var accountModel accountModels.AccountModel
	accountModel.Update(id, map[string]interface{}{"isActive": 1})
	c.JSON(200, gin.H{"message": "Activated"})
}

func (*SuperAdminHandlers) Unactive(c *gin.Context) {
	id := c.Param("id")
	var accountModel accountModels.AccountModel
	accountModel.Update(id, map[string]interface{}{"isActive": 0})
	c.JSON(200, gin.H{"message": "Deactivated"})
}

func (*SuperAdminHandlers) GetAccesses(c *gin.Context) {
	id := c.Param("id")
	var accessModel accessModels.AccessModel
	accesses, _ := accessModel.GetAll(map[string]interface{}{"accountId": id})
	c.JSON(200, accesses)
}

func (*SuperAdminHandlers) AddAccess(c *gin.Context) {
	id := c.Param("id")
	register := c.Query("register")
	if register == "" {
		c.JSON(400, gin.H{"error": "Register parameter required"})
		return
	}
	var accessModel accessModels.AccessModel
	err := accessModel.Create(accessTypes.Access{AccountId: id, Register: register})
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed or already exists"})
		return
	}
	c.JSON(200, gin.H{"message": "Access added"})
}

func (*SuperAdminHandlers) RemoveAccess(c *gin.Context) {
	id := c.Param("id")
	register := c.Query("register")
	var accessModel accessModels.AccessModel
	accessModel.Delete(accessTypes.Access{AccountId: id, Register: register})
	c.JSON(200, gin.H{"message": "Access removed"})
}
