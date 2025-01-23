package controller

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct{
	Email string
	jwt.StandardClaims
}

var jwtSecret []byte

func init(){
	jwtSecret=[]byte(os.Getenv("JWT_SECRET"))
}

func GenerateToken(email string) (string,error){
	expirationTime:=time.Now().Add(time.Hour*24*7)
	claims:=&Claims{
		Email: email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "codify",
		},
	}
	
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	tokenString,err:=token.SignedString(jwtSecret)
	if err!=nil{
		return "",err
	}
	return tokenString,nil
}

func VerifyToken(tokenString string) error{
	token,err:=jwt.Parse(tokenString,func(t *jwt.Token) (interface{}, error) {
		return jwtSecret,nil
	})
	if err!=nil{
		return err
	}
	if !token.Valid{
		return fmt.Errorf("invalid token")
	}
	return nil
}