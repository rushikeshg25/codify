package controller

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

type requestBody struct{
	Name string `json:"name"`
	Type string `json:"type"`
}
type codegroundType string

type codeground struct{
	id string
	userId int
	name string
	codeground_type codegroundType
}

type CodegroundController struct {
	db *sql.DB
}


const (
	REACT codegroundType = "REACT"
	NODE codegroundType = "NODE"
)



func NewCodegroundController(db *sql.DB) *CodegroundController {
	return &CodegroundController{
		db: db,
	}
}

func (q *CodegroundController) GetPlaygrounds(c *gin.Context) {
	var codegrounds []codeground
	var err error
	email,Exists:=c.Get("Email")
	if !Exists{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Internal server error"})
		return
	}
	err=q.db.Query("SELECT id,name,codeground_type FROM codegrounds WHERE user_id = ?", email).Scan(&codegrounds)


	c.JSON(200, gin.H{"message": "Hello"})
}

func (q *CodegroundController) CreatePlayground(c *gin.Context) {
	var reqbody requestBody
	var id int
	var err error
	email,Exists:=c.Get("Email")
	if !Exists{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Internal server error"})
		return
	}

	err=c.BindJSON(&reqbody)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Internal server error"})
		return
	}
	
	err=q.db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error":"Internal server error"})
		return	
	}
	data, err := q.db.Exec("INSERT INTO codegrounds(id,name,type,user_id) VALUES(?,?,?)", reqbody.Name, reqbody.Type, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	codegroundID, err := data.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"codeground_id": codegroundID})

}

func (q *CodegroundController) GetPlayground(c *gin.Context) {}

func (q *CodegroundController) UpdatePlayground(c *gin.Context) {}

func (q *CodegroundController) DeletePlayground(c *gin.Context) {}
