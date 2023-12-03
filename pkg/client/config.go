package client

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	User     string `env:"CLIENT"`         // users name
	UserID   string `env:"CLIENT"`         // this is generated at creation time. won't be in the initial .env file
	Email    string `env:"CLIENT_EMAIL"`   // users email
	Root     string `env:"CLIENT_ROOT"`    // client service root (ie. ../sfs/client/run/user-name)
	TestRoot string `env:"CLIENT_TESTING"` // testing root directory
	Port     int    `env:"CLIENT_PORT"`    // port for http client
}

func ClientConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode client config .env file: %s", err)
	}
	return &c
}
