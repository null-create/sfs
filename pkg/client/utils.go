package client

import (
	"log"
	"os"
	"strconv"

	svc "github.com/sfs/pkg/service"
)

func GetWd() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return wd
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
