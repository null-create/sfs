package server

import (
	"log"
	"time"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	Server serverConf
}

type serverConf struct {
	Host int    `env:"SERVER_HOST,required"`
	Port int    `env:"SERVER_PORT,required"`
	Addr string `env:"SERVER_ADDR,required"`

	Admin    string `env:"SERVER_ADMIN,required"`
	AdminKey string `env:"SERVER_ADMIN_KEY,required"`

	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
}

func SrvConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
