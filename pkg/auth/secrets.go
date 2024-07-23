package auth

import (
	"crypto/rand"
	"log"

	"golang.org/x/crypto/bcrypt"

	"github.com/sfs/pkg/env"
)

func GetSecret() ([]byte, error) {
	envCfg := env.NewE()
	if s, err := envCfg.Get("JWT_SECRET"); err == nil {
		return []byte(s), nil
	} else {
		return nil, err
	}
}

// generate a random string of n length to use as a secret
//
// technique from: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func GenSecret(length int) string {
	chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"
	charLen := len(chars)
	b := make([]byte, length)
	_, err := rand.Read(b) // generates len(b) random bytes
	if err != nil {
		log.Fatalf("failed to generate secret: %v", err)
	}
	for i := 0; i < length; i++ {
		b[i] = chars[int(b[i])%charLen]
	}
	return string(b)
}

// hash a given password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

// check if a given password is hashed correctly
func CheckPasswordHash(pwPlainText, hashedPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPw), []byte(pwPlainText))
	return err == nil
}
