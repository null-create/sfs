package client

import (
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	svc "github.com/sfs/pkg/service"
)

func GetWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
}

// Generate a random integer in the range [1, n)
func RandInt(limit int) int {
	// Seed the random number generator with the current time
	rand.Seed(time.Now().UnixNano())

	num := rand.Intn(limit)
	if num == 0 {
		return 1
	}
	return num
}

// parse the NEW_SERVICE env var to see if we are instantiating a new sfs service
func isMode(k string) bool {
	env := svc.NewE()
	v, err := env.Get(k)
	if err != nil {
		log.Fatalf("failed to get %s env var: %v", k, err)
	}
	isMode, err := strconv.ParseBool(v)
	if err != nil {
		log.Fatalf("failed to parse env var string to bool: %v", err)
	}
	return isMode
}

// copy a file
func Copy(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}
	return nil
}
