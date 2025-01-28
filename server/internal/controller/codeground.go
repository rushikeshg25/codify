package controller

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

type requestBody struct {
	Name string `json:"name"`
	Type string `json:"codeground_type"`
}
type codegroundType string

type codeground struct {
	id              string
	userId          int
	name            string
	codeground_type codegroundType
	createdAt       time.Time
	updatedAt       time.Time
}

type QueueMessage struct {
	Codeground *codeground `json:"codeground"`
	Msg        string      `json:"msg"` // Possible values: "start", "stop", "create", "update", "delete"
}

type CodegroundController struct {
	db *sql.DB      // Database connection
	ch *amqp.Channel // RabbitMQ channel
}

const (
	REACT codegroundType = "REACT"
	NODE  codegroundType = "NODE"
)


func NewCodegroundController(db *sql.DB) *CodegroundController {
	return &CodegroundController{
		db: db,
	}
}

func (q *CodegroundController) GetPlaygrounds(c *gin.Context) {
	var codegrounds []codeground
	var err error
	userId, Exists := c.Get("userId")
	
	if !Exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userId from context"})
		return
	}

	rows, err := q.db.Query("SELECT * FROM codegrounds WHERE user_id = ?", userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cg codeground
		if err := rows.Scan(&cg.id, &cg.userId, &cg.name, &cg.codeground_type, &cg.createdAt, &cg.updatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan row"})
			return
		}
		codegrounds = append(codegrounds, cg)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error iterating rows"})
		return
	}

	c.JSON(200, gin.H{"data": codegrounds})

}

func (q *CodegroundController) CreatePlayground(c *gin.Context) {
	var reqbody requestBody
	var err error
	userId, Exists := c.Get("userId")
	if !Exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	err = c.BindJSON(&reqbody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	data, err := q.db.Exec("INSERT INTO codegrounds(id,name,type,user_id) VALUES(?,?,?)", reqbody.Name, reqbody.Type, userId)
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

func (q *CodegroundController) GetPlayground(c *gin.Context) {
	codegroundId := c.Param("codegroundId")
	var codeground codeground

	if codegroundId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codegroundId is required"})
		return
	}

	err := q.db.QueryRow("SELECT * FROM codegrounds WHERE id = ?", codegroundId).
		Scan(&codeground.id, &codeground.userId, &codeground.name, &codeground.codeground_type, &codeground.createdAt, &codeground.updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Playground not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": codeground})
}

func (q *CodegroundController) UpdatePlayground(c *gin.Context) {
	var reqbody requestBody
	codegroundId := c.Param("codegroundId")
	var codeground codeground
	var err error

	if codegroundId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codegroundId is required"})
		return
	}
	err = q.db.QueryRow("SELECT * FROM codegrounds WHERE id = ?", codegroundId).
		Scan(&codeground.id, &codeground.userId, &codeground.name, &codeground.codeground_type, &codeground.createdAt, &codeground.updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Playground not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}
	err = c.BindJSON(&reqbody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	_, err = q.db.Exec("UPDATE codegrounds SET name = ?, type = ? WHERE id = ?", reqbody.Name, reqbody.Type, codegroundId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return

	}
	c.JSON(http.StatusOK, gin.H{"data": codeground})
}

func (q *CodegroundController) DeletePlayground(c *gin.Context) {
	codegroundId := c.Param("codegroundId")
	var err error

	if codegroundId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codegroundId is required"})
		return
	}
	_, err = q.db.Exec("DELETE FROM codegrounds WHERE id = ?", codegroundId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return

	}
	c.JSON(http.StatusOK, gin.H{"data": nil})
}
