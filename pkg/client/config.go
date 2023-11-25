package client

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	// users name
	User string `env:"CLIENT"`
	// users email
	Email string `env:"CLIENT_EMAIL"`
	// client service root
	// ie. ../sfs/client/run/user-name
	Root string `env:"CLIENT_ROOT"`
	// testing root directory
	TestRoot string `env:"CLIENT_TESTING"`
}

func ClientConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode client config .env file: %s", err)
	}
	return &c
}
