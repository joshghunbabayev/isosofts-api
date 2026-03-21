package superAdminRoutes

import (
	superAdminHandlers "isosofts-api/handlers/superAdmin"
	"isosofts-api/middlewares"

	"github.com/gin-gonic/gin"
)

func MainRoutes(rg *gin.RouterGroup) {
	var h superAdminHandlers.SuperAdminHandlers

	rg.POST("/login", h.Login)

	auth := rg.Group("")
	auth.Use(middlewares.SuperAdminAuthMiddleware())
	{
		auth.GET("/company", h.GetAllCompanies)
		auth.POST("/company", h.CreateCompany)
		auth.GET("/account", h.GetAllAccounts)

		company := auth.Group("/company/:companyId")
		{
			company.PUT("", h.UpdateCompany)
			company.GET("/account", h.GetCompanyAccounts)
			company.POST("/account", h.Create)
			company.PUT("/active", h.ActivateCompany)
			company.PUT("/unactive", h.DeactivateCompany)
		}

		user := auth.Group("/user/:id")
		{
			user.GET("", h.GetOne)
			user.PUT("", h.Update)
			user.PUT("/password", h.UpdatePassword)
			user.PUT("/lineManager", h.UpdateLineManager)
			user.PUT("/active", h.Active)
			user.PUT("/unactive", h.Unactive)

			user.GET("/access", h.GetAccesses)
			user.POST("/access", h.AddAccess)
			user.DELETE("/access", h.RemoveAccess)
		}
	}
}
