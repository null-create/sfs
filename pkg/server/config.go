package server

import (
	"log"
	"time"

	"github.com/joeshaw/envdecode"
)

type SvrCnf struct {
	Port int    `env:"SERVER_PORT,required"`
	Addr string `env:"SERVER_ADDR,required"`

	Admin    string `env:"SERVER_ADMIN,required"`
	AdminKey string `env:"SERVER_ADMIN_KEY,required"`

	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
}

func ServerConfig() *SvrCnf {
	var c SvrCnf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode server config .env file: %s", err)
	}
	return &c
}

type SvcCfg struct {
	SvcRoot string `env:"SERVICE_ROOT,required"`
}

func ServiceConfig() *SvcCfg {
	var c SvcCfg
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode server config .env file: %s", err)
	}
	return &c
}
