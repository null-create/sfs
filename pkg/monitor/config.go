package monitor

import (
	"log"

	"github.com/sfs/pkg/configs"

	"github.com/joeshaw/envdecode"
)

type MonitorConfigs struct {
	Buffered bool `env:"BUFFERED_EVENTS,required"`
	BuffSize int  `env:"EVENT_BUFFER_SIZE,requred"`
}

func MonitorCfgs() *MonitorConfigs {
	configs.SetEnv(false)

	var c MonitorConfigs
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("failed to decode monitor config .env file: %s", err)
	}
	return &c
}

var MonCfgs = MonitorCfgs()
