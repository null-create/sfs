package client

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	IsAdmin   bool   `env:"ADMIN_MODE"`               // whether the service should be run in admin mode or not
	User      string `env:"CLIENT,required"`          // users name
	UserAlias string `env:"CLIENT_USERNAME,required"` // users alias (username)
	UserID    string `env:"CLIENT_ID,required"`       // this is generated at creation time. won't be in the initial .env file
	Email     string `env:"CLIENT_EMAIL,required"`    // users email
	Root      string `env:"CLIENT_ROOT,required"`     // client service root (ie. ../sfs/client/run/)
	TestRoot  string `env:"CLIENT_TESTING,required"`  // testing root directory
	Port      int    `env:"CLIENT_PORT,required"`     // port for http client
	Addr      string `env:"CLIENT_ADDRESS,required"`  // address for http client
}

func ClientConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode client config .env file: %s", err)
	}
	return &c
}

// client env, user, and service configurations
var cfgs = ClientConfig()
