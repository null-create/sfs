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
	// client service root, ie ../sfs/client/root
	//
	// all users file will be uder .../sfs/root/user-name/root
	Root string `env:"CLIENT_ROOT"`
}

func ClientConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode client config .env file: %s", err)
	}
	return &c
}
