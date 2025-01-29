package middleware

import (
	"server/internal/controller"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) {
    if c.Request.Method == "OPTIONS" {
        c.Next()
        return
    }
    
    token, err := c.Cookie("token")
    if err != nil || token == "" || controller.VerifyToken(token) != nil {
        c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
        return
    }
    _, exists := c.Get("Email")
    if !exists {
        userData, err := controller.GetUserDataFromToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
            return
        }
        
        c.Set("userId", userData.Id)
        c.Set("email", userData.Email)
    }
    c.Next()
}