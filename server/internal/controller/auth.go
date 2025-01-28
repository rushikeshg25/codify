package controller

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	db *sql.DB
}

type authRequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{
		db: db,
	}
}

func (q *AuthController) Login(c *gin.Context) {
	var reqBody authRequestBody
	var email string
	var id int
	var password string
	var createdAt []uint8
	
	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization failed"})
		return
	}
	err := q.db.QueryRow("SELECT * FROM users WHERE email = ?", reqBody.Email).Scan(&id, &email, &password, &createdAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	if !CheckHashPassword(reqBody.Password, password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	token, err := GenerateToken(email, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.SetCookie("token", token, 3600*24*7, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}

func (q *AuthController) Signup(c *gin.Context) {
	var reqBody authRequestBody
	var email string
	var err error
	if err = c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = q.db.QueryRow("SELECT id FROM users WHERE email = ?", reqBody.Email).Scan(&email)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	hash, err := HashPassword(reqBody.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	data, err := q.db.Exec("INSERT INTO users (email,password) VALUES (?,?)", reqBody.Email, hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	userID, err := data.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	token, err := GenerateToken(email, int(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	c.SetCookie("token", token, 3600*24*7, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful", "Email": reqBody.Email})
}

func (q *AuthController) Logout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckHashPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
