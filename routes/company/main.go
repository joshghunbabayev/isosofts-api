package companyRoutes

import (
	companyHandlers "isosofts-api/handlers/company"
	"isosofts-api/middlewares"

	"github.com/gin-gonic/gin"
)

func MainRoutes(rg *gin.RouterGroup) {
	var companyHandler companyHandlers.CompanyHandlers

	selfRoutes := rg.Group("/self")
	selfRoutes.Use(middlewares.AuthMiddleware())
	{
		selfRoutes.GET("", companyHandler.Get)

		adminRoutes := selfRoutes.Group("")
		adminRoutes.Use(middlewares.AdminAuthMiddleware())
		{
			adminRoutes.PUT("", companyHandler.Update)
		}
	}
}
