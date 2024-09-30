package auth

import (
	"crypto/rand"
	"log"
	"math/big"

	"github.com/sfs/pkg/configs"

	"golang.org/x/crypto/bcrypt"
)

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-"

func GetSecret() ([]byte, error) {
	s, err := configs.NewSvcConfig().Get(configs.JWT_SECRET)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

// generate a random string of n length to use as a secret
//
// technique from: https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func GenSecret(length int) string {
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
	b, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(b), err
}

// check if a given password is hashed correctly
func CheckPasswordHash(pwPlainText, hashedPw string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPw), []byte(pwPlainText))
	return err == nil
}

// technique from: https://github.com/a-h/templ/blob/main/examples/content-security-policy/main.go
func GenNonce(size int) (string, error) {
	ret := make([]byte, size)
	for i := 0; i < size; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		ret[i] = chars[num.Int64()]
	}
	return string(ret), nil
}
