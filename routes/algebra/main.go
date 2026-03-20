package algebraRoutes

import (
	algebraHandlers "isosofts-api/handlers/algebra"
	"isosofts-api/middlewares"

	"github.com/gin-gonic/gin"
)

func MainRoutes(rg *gin.RouterGroup) {
	var algebraHandler algebraHandlers.AlgebraHandlers

	routes := rg.Group("")
	routes.Use(middlewares.AuthMiddleware())
	{
		routes.GET("/self", algebraHandler.GetSelf)
		routes.GET("/check-access", algebraHandler.CheckAccess)
	}
}
