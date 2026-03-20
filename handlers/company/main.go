package companyHandlers

import (
	companyModels "isosofts-api/models/company"
	accountTypes "isosofts-api/types/account"
	"strings"

	"github.com/gin-gonic/gin"
)

type CompanyHandlers struct{}

func (*CompanyHandlers) Get(c *gin.Context) {
	account := c.MustGet("account").(accountTypes.Account)

	var companyModel companyModels.CompanyModel
	company, err := companyModel.GetById(account.CompanyId)

	if err != nil || company.IsEmpty() {
		c.JSON(404, gin.H{"error": "Company not found"})
		return
	}

	c.IndentedJSON(200, company)
}

func (*CompanyHandlers) Update(c *gin.Context) {
	admin := c.MustGet("account").(accountTypes.Account)

	var body struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON format"})
		return
	}

	errs := make(map[string]string)
	if strings.TrimSpace(body.Name) == "" {
		errs["name"] = "Company name is required"
	}

	if len(errs) > 0 {
		c.IndentedJSON(400, gin.H{"errors": errs})
		return
	}

	var companyModel companyModels.CompanyModel

	err := companyModel.Update(admin.CompanyId, map[string]interface{}{
		"name": body.Name,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update company name"})
		return
	}

	c.JSON(200, gin.H{"message": "Company name updated successfully"})
}
