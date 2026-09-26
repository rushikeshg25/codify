package controller

import (
	"database/sql"
	"net/http"
	"server/internal/queue"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jaevor/go-nanoid"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyz"
)

type requestBody struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"codeground_type" binding:"required"`
}
type codegroundType string

// isValidType reports whether t is one of the supported codeground types.
func isValidType(t string) bool {
	return codegroundType(t) == REACT || codegroundType(t) == NODE
}

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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read codegrounds"})
			return
		}
		codegrounds = append(codegrounds, cg)
	}

	c.JSON(http.StatusOK, gin.H{"data": codegrounds})
}

func (q *CodegroundController) CreateCodeground(c *gin.Context) {
	var reqbody requestBody
	var err error

	generator, err := nanoid.CustomUnicode(alphabet, 8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	userId, Exists := c.Get("userId")
	if !Exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if err = c.ShouldBindJSON(&reqbody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and codeground_type are required"})
		return
	}
	if !isValidType(reqbody.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codeground_type must be REACT or NODE"})
		return
	}

	codegroundID := generator()
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
	// Mint the codeground-scoped token up front so the browser never has to
	// talk to the pod unauthenticated.
	accessToken, err := GenerateCodegroundToken(codegroundID, userId.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": codegroundID, "name": reqbody.Name, "codeground_type": reqbody.Type, "user_id": userId, "access_token": accessToken, "created_at": time.Now().Format("2006-01-02 15:04:05"), "updated_at": time.Now().Format("2006-01-02 15:04:05")})

}

func (q *CodegroundController) GetCodeground(c *gin.Context) {
	codegroundId := c.Param("codegroundId")
	var codeground codeground

	if codegroundId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codegroundId is required"})
		return
	}
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userId from context"})
		return
	}

	err := q.db.QueryRow("SELECT id, user_id, name, codeground_type, created_at, updated_at FROM codegrounds WHERE id = ? AND user_id = ?", codegroundId, userId).
		Scan(&codeground.id, &codeground.userId, &codeground.name, &codeground.codeground_type, &codeground.createdAt, &codeground.updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Codeground not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// The ownership check above (id AND user_id) is what authorises this
	// token, so it can only ever be issued to the codeground's owner.
	accessToken, err := GenerateCodegroundToken(codeground.id, codeground.userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"name": codeground.name, "id": codeground.id, "codeground_type": codeground.codeground_type, "access_token": accessToken})
}

// GetCodegroundToken re-issues a codeground access token. The browser calls
// this when its current token is close to expiring.
func (q *CodegroundController) GetCodegroundToken(c *gin.Context) {
	codegroundId := c.Param("codegroundId")
	if codegroundId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codegroundId is required"})
		return
	}
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userId from context"})
		return
	}

	var ownerId int
	err := q.db.QueryRow("SELECT user_id FROM codegrounds WHERE id = ? AND user_id = ?", codegroundId, userId).Scan(&ownerId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Codeground not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	accessToken, err := GenerateCodegroundToken(codegroundId, ownerId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
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
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userId from context"})
		return
	}
	err = q.db.QueryRow("SELECT id, user_id, name, codeground_type, created_at, updated_at FROM codegrounds WHERE id = ? AND user_id = ?", codegroundId, userId).
		Scan(&codeground.id, &codeground.userId, &codeground.name, &codeground.codeground_type, &codeground.createdAt, &codeground.updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Codeground not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}
	if err = c.ShouldBindJSON(&reqbody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name and codeground_type are required"})
		return
	}
	if !isValidType(reqbody.Type) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codeground_type must be REACT or NODE"})
		return
	}
	_, err = q.db.Exec("UPDATE codegrounds SET name = ?, codeground_type = ? WHERE id = ? AND user_id = ?", reqbody.Name, reqbody.Type, codegroundId, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return

	}
	// Respond with the updated values rather than the pre-update row.
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"id":              codeground.id,
		"user_id":         codeground.userId,
		"name":            reqbody.Name,
		"codeground_type": reqbody.Type,
	}})
}

func (q *CodegroundController) DeleteCodeground(c *gin.Context) {
	codegroundId := c.Param("codegroundId")
	var err error

	if codegroundId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "codegroundId is required"})
		return
	}
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get userId from context"})
		return
	}
	res, err := q.db.Exec("DELETE FROM codegrounds WHERE id = ? AND user_id = ?", codegroundId, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Codeground not found"})
		return
	}
	q.q.PublishToQueue(queue.Codeground{
		ID:     codegroundId,
		UserID: userId.(int),
	}, "delete")
	c.JSON(http.StatusOK, gin.H{"data": nil})
}
