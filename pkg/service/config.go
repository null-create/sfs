package service

import (
	"fmt"
	"log"

	"github.com/joeshaw/envdecode"
)

// server's root endpoint, since we're running on LAN
const Endpoint = "http://localhost"

type ServiceConfig struct {
	Port string `env:"SERVER_PORT,required"`
}

func NewSvcCfg() *ServiceConfig {
	var cfg ServiceConfig
	if err := envdecode.StrictDecode(&cfg); err != nil {
		log.Fatalf(fmt.Sprintf("failed to decode service config: %v", err))
	}
	return &cfg
}
