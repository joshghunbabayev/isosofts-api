package middlewares

import (
	accountModels "isosofts-api/models/account"
	accountTypes "isosofts-api/types/account"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token is required"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			return
		}

		claims, _ := token.Claims.(jwt.MapClaims)
		accountId := claims["sub"].(string)

		var accountModel accountModels.AccountModel
		account, err := accountModel.GetById(accountId)

		if err != nil || account.IsEmpty() { //
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		if account.IsActive == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Your account is deactivated"})
			return
		}

		c.Set("account", account)
		c.Next()
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		account := c.MustGet("account").(accountTypes.Account)

		if account.IsAdmin != 1 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Only admins can access this route",
			})
			return
		}

		c.Next()
	}
}

func SuperAdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.Query("token")

		if tokenString == "" {
			c.AbortWithStatusJSON(401, gin.H{"error": "Token is required"})
			return
		}

		jwtSecret := os.Getenv("JWT_SECRET_FOR_ADMIN")
		adminId := os.Getenv("ADMIN_ID")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid or expired admin token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["sub"] != adminId {
			c.AbortWithStatusJSON(403, gin.H{"error": "Access denied: Not a super admin"})
			return
		}

		c.Next()
	}
}
