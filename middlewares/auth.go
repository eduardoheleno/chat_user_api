package middlewares

import (
	"jwt_auth"

	"github.com/gin-gonic/gin"
)

func ProtectRoute() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("authToken")
		if len(token) <= 0 {
			c.AbortWithStatusJSON(401, gin.H{"error": "User not authenticated"})
			return
		}

		claims := jwt_auth.ValidateToken(token)
		if claims == nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "User not authenticated"})
			return
		}

		c.Set("userId", claims.UserId)
		c.Set("authToken", token)
		c.Next()
	}
}
