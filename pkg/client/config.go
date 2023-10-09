package client

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	Root string `env:"CLIENT_ROOT"`
}

func ClientConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode client config .env file: %s", err)
	}
	return &c
}
