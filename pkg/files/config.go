package files

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

func ServerConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode server config .env file: %s", err)
	}
	return &c
}

type SFSConf struct {
	// physical location of the the sfs root
	// service directory on the server
	ServiceRoot string `env:"SERVICE_ROOT,required"`
}

func ServiceConfig() *SFSConf {
	var c SFSConf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode service .env file: %s", err)
	}
	return &c
}
