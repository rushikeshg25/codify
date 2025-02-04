package controller

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"server/internal/queue"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaevor/go-nanoid"
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
	db *sql.DB // Database connection
	q  *queue.Queue
}

const (
	REACT codegroundType = "REACT"
	NODE  codegroundType = "NODE"
)

func NewCodegroundController(db *sql.DB, queue *queue.Queue) *CodegroundController {
	return &CodegroundController{
		db: db,
		q:  queue,
	}
}

func (q *CodegroundController) GetCodegrounds(c *gin.Context) {
	var codegrounds []struct {
		ID             string         `json:"id"`
		UserID         int            `json:"userId"`
		Name           string         `json:"name"`
		CodegroundType codegroundType `json:"codeground_type"`
		CreatedAt      string         `json:"createdAt"`
		UpdatedAt      string         `json:"updatedAt"`
	}

	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userId from context"})
		return
	}

	rows, err := q.db.Query("SELECT id, user_id, name, codeground_type, created_at, updated_at FROM codegrounds WHERE user_id = ?", userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var cg struct {
			ID             string         `json:"id"`
			UserID         int            `json:"userId"`
			Name           string         `json:"name"`
			CodegroundType codegroundType `json:"codeground_type"`
			CreatedAt      string         `json:"createdAt"`
			UpdatedAt      string         `json:"updatedAt"`
		}
		err := rows.Scan(&cg.ID, &cg.UserID, &cg.Name, &cg.CodegroundType, &cg.CreatedAt, &cg.UpdatedAt)
		if err != nil {
			log.Fatal("Row scan error:", err)
		}
		codegrounds = append(codegrounds, cg)
	}

	c.JSON(http.StatusOK, gin.H{"data": codegrounds})
}

func (q *CodegroundController) CreateCodeground(c *gin.Context) {
	var reqbody requestBody
	var err error
	generator,err := nanoid.Standard(8)
	if err != nil {
		log.Fatalf("Error generating ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
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

	codegroundID:=generator()
	if err != nil {
		log.Fatalf("Error generating ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	_, err = q.db.Exec(
		"INSERT INTO codegrounds(id, name, codeground_type, user_id) VALUES(?, ?, ?, ?)",
		codegroundID,
		reqbody.Name,
		reqbody.Type,
		userId,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	q.q.PublishToQueue(queue.Codeground{
		ID:             codegroundID,
		UserID:         userId.(int),
		Name:           reqbody.Name,
		CodegroundType: reqbody.Type,
		CreatedAt:      time.Now().Format("2006-01-02 15:04:05"),
		UpdatedAt:      time.Now().Format("2006-01-02 15:04:05"),
	}, "create")
	c.JSON(http.StatusCreated, gin.H{"id": codegroundID, "name": reqbody.Name, "codeground_type": reqbody.Type, "user_id": userId, "created_at": time.Now().Format("2006-01-02 15:04:05"), "updated_at": time.Now().Format("2006-01-02 15:04:05")})

}

func (q *CodegroundController) GetCodeground(c *gin.Context) {
	codegroundId := c.Param("codegroundId")
	var codeground codeground

	if codegroundId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codegroundId is required"})
		return
	}

	row := q.db.QueryRow("SELECT * FROM codegrounds WHERE id = ?", codegroundId).
		Scan(&codeground.id, &codeground.userId, &codeground.name, &codeground.codeground_type, &codeground.createdAt, &codeground.updatedAt)
	fmt.Println(codeground)
	if row == nil {
		c.JSON(200, gin.H{"error": "No Codegrounds"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"name": codeground.name, "id": codeground.id, "codeground_type": codeground.codeground_type})
}

func (q *CodegroundController) UpdateCodeground(c *gin.Context) {
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
			c.JSON(http.StatusNotFound, gin.H{"error": "Codeground not found"})
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

func (q *CodegroundController) DeleteCodeground(c *gin.Context) {
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
