package controller

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

type PlaygroundController struct {
	db *sql.DB
}

func NewPlaygroundController(db *sql.DB) *PlaygroundController {
	return &PlaygroundController{
		db: db,
	}
}

func (q *PlaygroundController) GetPlaygrounds(c *gin.Context) {
	val, _ := c.Get("Email")

	c.JSON(200, gin.H{"message": "Hello"})
}
