package middleware

import (
	"server/internal/controller"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
	var err error
	token, err := c.Cookie("token")

	if err != nil || token == "" || controller.VerifyToken(token) != nil {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		c.Abort()
		return
	}
	_, exists := c.Get("Email")
	if !exists {
		userData, err := controller.GetUserDataFromToken(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Set("userId", userData.Id)
		c.Set("email", userData.Email)
	}
	c.Next()
}
