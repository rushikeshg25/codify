package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

// codegroundSecret is the master key that per-codeground signing keys are
// derived from. It is deliberately separate from jwtSecret: a codeground pod
// is handed a derived key, and must never be able to mint session tokens.
var codegroundSecret []byte

func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET must be set: refusing to sign tokens with an empty key")
	}
	jwtSecret = []byte(secret)

	cgSecret := os.Getenv("CODEGROUND_TOKEN_SECRET")
	if cgSecret == "" {
		log.Fatal("CODEGROUND_TOKEN_SECRET must be set: refusing to sign codeground tokens with an empty key")
	}
	codegroundSecret = []byte(cgSecret)
}

// CodegroundKey derives the signing key for a single codeground.
//
// Every pod receives only its own derived key, so a pod running hostile user
// code cannot forge access tokens for anyone else's codeground. The manager
// performs this identical derivation when it injects the key into the pod, so
// the two must stay in sync.
func CodegroundKey(codegroundID string) []byte {
	mac := hmac.New(sha256.New, codegroundSecret)
	mac.Write([]byte("codeground:" + codegroundID))
	return []byte(hex.EncodeToString(mac.Sum(nil)))
}

// codegroundTokenTTL is how long a codeground access token stays valid. The
// browser re-fetches one whenever it loads the codeground, so this can be
// short without hurting usability.
func codegroundTokenTTL() time.Duration {
	if raw := os.Getenv("CODEGROUND_TOKEN_TTL"); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil {
			return d
		}
		log.Printf("invalid CODEGROUND_TOKEN_TTL %q, falling back to 1h", raw)
	}
	return time.Hour
}

// CodegroundClaims scope a token to exactly one codeground.
type CodegroundClaims struct {
	CodegroundId string `json:"codegroundId"`
	UserId       int    `json:"userId"`
	jwt.StandardClaims
}

// GenerateCodegroundToken mints a short-lived token that authorises the
// terminal and file APIs of a single codeground.
func GenerateCodegroundToken(codegroundID string, userID int) (string, error) {
	now := time.Now()
	claims := &CodegroundClaims{
		CodegroundId: codegroundID,
		UserId:       userID,
		StandardClaims: jwt.StandardClaims{
			Subject:   codegroundID,
			Audience:  "codeground",
			ExpiresAt: now.Add(codegroundTokenTTL()).Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    "codify",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(CodegroundKey(codegroundID))
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
