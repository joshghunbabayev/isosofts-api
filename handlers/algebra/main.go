package algebraHandlers

import (
	accessModels "isosofts-api/models/access"
	accountTypes "isosofts-api/types/account"
	"strings"

	"github.com/gin-gonic/gin"
)

type AlgebraHandlers struct{}

// 1. Identity Handler: Token-ə əsasən hesabı qaytarır
func (*AlgebraHandlers) GetSelf(c *gin.Context) {
	// AuthMiddleware artıq hesabı kontekstə yerləşdirib
	account, _ := c.MustGet("account").(accountTypes.Account)

	// Uğurlu sorğu - 200
	c.JSON(200, account)
}

// 2. Access Check Handler: Token və Register-ə görə icazəni yoxlayır
func (*AlgebraHandlers) CheckAccess(c *gin.Context) {
	account, _ := c.MustGet("account").(accountTypes.Account)
	register := c.Query("register")

	if strings.TrimSpace(register) == "" {
		c.JSON(400, gin.H{"error": "Register parameter is required"})
		return
	}

	var accessModel accessModels.AccessModel
	// Bazada bu accountId və register kombinasiyasını axtarırıq
	accesses, err := accessModel.GetAll(map[string]interface{}{
		"accountId": account.Id,
		"register":  register,
	})

	if err != nil {
		c.JSON(500, gin.H{"error": "Internal database error"})
		return
	}

	// Əgər nəticə varsa (len > 0), girişə icazə verilir
	if len(accesses) > 0 {
		c.JSON(200, gin.H{
			"hasAccess": true,
			"accountId": account.Id,
			"register":  register,
		})
	} else {
		// Giriş qadağandırsa - 403 (Forbidden)
		c.JSON(403, gin.H{
			"hasAccess": false,
			"error":     "No access for this register",
		})
	}
}
