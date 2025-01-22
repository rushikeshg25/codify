package controller

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	db *sql.DB
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{
		db: db,
	}
}

func (q* AuthController) Login(c *gin.Context) {
	
}


func (q* AuthController) Signup(c *gin.Context) {}