package logger

import (
	"log"

	"github.com/joeshaw/envdecode"
	"github.com/sfs/pkg/env"
)

type Conf struct {
	LogDir string `env:"SERVICE_LOG_DIR,required"`
}

func LogConfig() *Conf {
	env.SetEnv(false)

	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode log config .env file: %s", err)
	}
	return &c
}

// import logging configurations
var logCfg = LogConfig()
