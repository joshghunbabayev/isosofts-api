package accountRoutes

import (
	accountHandlers "isosofts-api/handlers/account"
	"isosofts-api/middlewares"

	"github.com/gin-gonic/gin"
)

func MainRoutes(rg *gin.RouterGroup) {
	var accountHandler accountHandlers.AccountHandlers
	rg.POST("/signup", accountHandler.SignUp)
	rg.POST("/login", accountHandler.Login)

	routes := rg.Group("")
	routes.Use(middlewares.AuthMiddleware())
	{
		selfRoutes := routes.Group("/self")
		{
			selfRoutes.GET("", accountHandler.GetSelf)
			selfRoutes.PUT("/password", accountHandler.UpdatePasswordSelf)
		}

		staffRoutes := routes.Group("/staff")
		{
			staffRoutes.GET("", accountHandler.GetAll)
			staffRoutes.GET("/:id", accountHandler.GetOne)

			staffForAdminRoutes := staffRoutes.Group("")
			staffForAdminRoutes.Use(middlewares.AdminAuthMiddleware())

			staffForAdminRoutes.POST("", accountHandler.Create)
			staffForAdminRoutes.PUT("/:id", accountHandler.Update)
			staffForAdminRoutes.PUT("/:id/lineManager", accountHandler.UpdateLineManager)
			staffForAdminRoutes.PUT("/:id/password", accountHandler.UpdatePassword)
			staffForAdminRoutes.PUT("/:id/unactive", accountHandler.Unactive)
			staffForAdminRoutes.PUT("/:id/active", accountHandler.Active)
			staffForAdminRoutes.GET("/:id/access", accountHandler.GetAccesses)
			staffForAdminRoutes.POST("/:id/access", accountHandler.AddAccess)
			staffForAdminRoutes.DELETE("/:id/access", accountHandler.RemoveAccess)
		}
	}
}
