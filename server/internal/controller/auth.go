package controller

import (
	"database/sql"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// cookieSecure reports whether the auth cookie should carry the Secure flag.
// It is off by default for local HTTP dev and enabled via COOKIE_SECURE=true
// in any environment served over HTTPS.
func cookieSecure() bool {
	return os.Getenv("COOKIE_SECURE") == "true"
}

// setAuthCookie writes the JWT as an HttpOnly, SameSite=Lax cookie.
func setAuthCookie(c *gin.Context, token string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("token", token, maxAge, "/", "", cookieSecure(), true)
}

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
	err := q.db.QueryRow("SELECT id, email, password, created_at FROM users WHERE email = ?", reqBody.Email).Scan(&id, &email, &password, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return the same response as a wrong password so we don't leak
			// which emails are registered (user enumeration).
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
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
	setAuthCookie(c, token, 3600*24*7)
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
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	} else if err != sql.ErrNoRows {
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
	token, err := GenerateToken(reqBody.Email, int(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	setAuthCookie(c, token, 3600*24*7)

	c.JSON(http.StatusOK, gin.H{"message": "Signup successful", "Email": reqBody.Email})
}

func (q *AuthController) Logout(c *gin.Context) {
	setAuthCookie(c, "", -1)
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
