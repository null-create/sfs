package server

import (
	"log"
	"time"

	"github.com/sfs/pkg/env"

	"github.com/joeshaw/envdecode"
)

type SvrCnf struct {
	Port         int           `env:"SERVER_PORT,required"`
	Addr         string        `env:"SERVER_ADDR,required"`
	Admin        string        `env:"SERVER_ADMIN,required"`
	AdminKey     string        `env:"SERVER_ADMIN_KEY,required"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
}

func ServerConfig() *SvrCnf {
	env.SetEnv(false)

	var c SvrCnf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode server config .env file: %s", err)
	}
	return &c
}

var svrCfg = ServerConfig()

type SvcCfg struct {
	SvcRoot    string `env:"SERVICE_ROOT,required"`
	NewService bool   `env:"NEW_SERVICE,required"`
	IsAdmin    bool   `env:"ADMIN_MODE,required"`
}

func ServiceConfig() *SvcCfg {
	env.SetEnv(false)

	var c SvcCfg
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode server config .env file: %s", err)
	}
	return &c
}

var svcCfg = ServiceConfig()

// TODO: update and manage server and service configs
