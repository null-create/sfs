package server

import (
	"log"
	"time"

	"github.com/sfs/pkg/configs"

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
	configs.SetEnv(false)

	var c SvrCnf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("failed to decode server config .env file: %s", err)
	}
	return &c
}

var (
	svrCfg = ServerConfig()
	svcCfg = ServiceConfig()
)

type SvcCfg struct {
	EnvFile    string `env:"SERVICE_ENV,required"` // absoloute path to the dedicated .env file
	SvcRoot    string `env:"SERVICE_ROOT,required"`
	NewService bool   `env:"NEW_SERVICE,required"`
	IsAdmin    bool   `env:"ADMIN_MODE,required"`
}

func ServiceConfig() *SvcCfg {
	configs.SetEnv(false)

	var c SvcCfg
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("failed to decode server config .env file: %s", err)
	}
	return &c
}

// TODO: update and manage server and service configs
