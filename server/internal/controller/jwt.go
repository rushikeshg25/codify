package controller

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserId int    `json:"userId"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

type userData struct {
	Id    int
	Email string
}

var jwtSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET must be set: refusing to sign tokens with an empty key")
	}
	jwtSecret = []byte(secret)
}

// keyFunc returns the HMAC secret only when the token was signed with an HMAC
// algorithm. This pins the expected signing method and guards against
// algorithm-confusion attacks (e.g. a token forged with alg=none or RS256).
func keyFunc(t *jwt.Token) (any, error) {
	if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
	}
	return jwtSecret, nil
}

func GenerateToken(email string, userId int) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 24 * 7)
	claims := &Claims{
		Email:  email,
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "codify",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, keyFunc)
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func GetUserDataFromToken(tokenString string) (userData, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		return userData{}, err
	}
	if !token.Valid {
		return userData{}, fmt.Errorf("invalid token")
	}
	return userData{
		Id:    claims.UserId,
		Email: claims.Email,
	}, nil
}
