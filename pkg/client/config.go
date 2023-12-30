package client

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	User     string `env:"CLIENT,required"`         // users name
	UserID   string `env:"CLIENT_ID,required"`      // this is generated at creation time. won't be in the initial .env file
	Email    string `env:"CLIENT_EMAIL,required"`   // users email
	Root     string `env:"CLIENT_ROOT,required"`    // client service root (ie. ../sfs/client/run/)
	TestRoot string `env:"CLIENT_TESTING,required"` // testing root directory
	Port     int    `env:"CLIENT_PORT,required"`    // port for http client
	Addr     string `env:"CLIENT_ADDRESS,required"` // address for http client
}

func ClientConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode client config .env file: %s", err)
	}
	return &c
}
