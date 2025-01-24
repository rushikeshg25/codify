package controller

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

type CodegroundController struct {
	db *sql.DB
}

func NewCodegroundController(db *sql.DB) *CodegroundController {
	return &CodegroundController{
		db: db,
	}
}

func (q *CodegroundController) GetPlaygrounds(c *gin.Context) {
	val, _ := c.Get("Email")

	c.JSON(200, gin.H{"message": "Hello"})
}

func (q *CodegroundController) CreatePlayground(c *gin.Context) {}

func (q *CodegroundController) GetPlayground(c *gin.Context) {}

func (q *CodegroundController) UpdatePlayground(c *gin.Context) {}

func (q *CodegroundController) DeletePlayground(c *gin.Context) {}