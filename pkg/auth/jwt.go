package auth

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

// json web token
type Token struct {
	Jwt    string // token string
	Secret []byte // secret key
}

func NewT() *Token {
	s, err := getSecret()
	if err != nil {
		log.Fatalf("unable to retrieve token secret: %v", err)
	}
	return &Token{
		Secret: s,
	}
}

func getSecret() ([]byte, error) {
	e := NewE()
	if s, err := e.Get("JWT_SECRET"); err == nil {
		return []byte(s), nil
	} else {
		return nil, err
	}
}

// retrieve jwt token from request
func (t *Token) Extract(rawReqToken string) (string, error) {
	splitToken := strings.Split(rawReqToken, "Bearer ")
	if len(splitToken) != 2 { // bearer token not in proper format
		return "", fmt.Errorf("invalid token format")
	}
	reqToken := strings.TrimSpace(splitToken[1])
	return reqToken, nil
}

// verify jwt token and attempt ot retrieve userID from it.
//
// use the return value to compare against the db and whether
// they're an actual user
func (t *Token) Verify(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return t.Secret, nil
	})
	if err != nil {
		return "", err
	}
	if !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("failed to parse jwt claims")
	}
	userID := claims["sub"].(string)
	if userID == "" {
		return "", fmt.Errorf("no user ID found in token claims")
	}
	return userID, nil
}

// create a new token using a given payload/string
func (t *Token) Create(payload string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = payload
	claims["exp"] = time.Now().Add(time.Hour).UTC() // Token expires in 1  hour
	tokenString, err := token.SignedString(t.Secret)
	if err != nil {
		return "", err
	}
	t.Jwt = tokenString
	return tokenString, nil
}
