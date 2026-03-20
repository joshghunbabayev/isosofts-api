package routes

import (
	accountRoutes "isosofts-api/routes/account"
	algebraRoutes "isosofts-api/routes/algebra"
	companyRoutes "isosofts-api/routes/company"
	superAdminRoutes "isosofts-api/routes/superAdmin"

	"github.com/gin-gonic/gin"
)

func APIRoutes(rg *gin.RouterGroup) {
	accountRoutes.MainRoutes(rg.Group("/account"))
	companyRoutes.MainRoutes(rg.Group("/company"))
	algebraRoutes.MainRoutes(rg.Group("/algebra"))
	superAdminRoutes.MainRoutes(rg.Group("/superAdmin"))
}
