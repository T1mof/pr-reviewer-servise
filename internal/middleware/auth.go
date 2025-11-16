package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminAuth проверяет X-Admin-Token header.
func AdminAuth(adminToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Admin-Token")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "X-Admin-Token header required",
				},
			})
			c.Abort()
			return
		}

		if token != adminToken {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "invalid admin token",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
