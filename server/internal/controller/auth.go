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

type RequestBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{
		db: db,
	}
}


func (q* AuthController) Login(c *gin.Context) {
	var reqBody RequestBody
	var email string
	var id int
	var password string
	if err := c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err:= q.db.QueryRow("SELECT (id,email,password) FROM users WHERE email = ? AND password = ?", reqBody.Email,reqBody.Password).Scan(&id,&email,&password)
	if err != sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized,gin.H{"error": "Invalid credentials"})
		return
	}
	if !CheckHashPassword(reqBody.Password,password){
		c.JSON(http.StatusUnauthorized,gin.H{"error": "Invalid credentials"})
		return
	}
	
	token,err:=GenerateToken(email)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error": "Internal server error"})
		return
	}
	c.SetCookie("token", token, 3600*24*7, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
}


func (q* AuthController) Signup(c *gin.Context) {
	var reqBody RequestBody
	var email string
	var err error
	if err = c.BindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err= q.db.QueryRow("SELECT id FROM users WHERE email = ?", reqBody.Email).Scan(&email)
	if err!=sql.ErrNoRows{
		c.JSON(http.StatusInternalServerError,gin.H{"error": "Internal server error"})
		return
	}
	hash,err:=HashPassword(reqBody.Password)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error": "Internal server error"})
		return
	}
	_,err=q.db.Exec("INSERT INTO users (email,password) VALUES (?,?)",reqBody.Email,hash)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error": "Internal server error"})
		return
	}

	token,err:=GenerateToken(email)
	if err!=nil{
		c.JSON(http.StatusInternalServerError,gin.H{"error": "Internal server error"})
		return
	}
	c.SetCookie("token", token, 3600*24*7, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful"})
}

func HashPassword(password string) (string,error){
	bytes,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	return string(bytes),err
}

func CheckHashPassword(password string,hash string) bool{
	err:=bcrypt.CompareHashAndPassword([]byte(hash),[]byte(password))
	return err==nil
}