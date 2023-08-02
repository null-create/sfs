package pkg

import (
	"log"

	"github.com/joeshaw/envdecode"
)

type NimbusConf struct {
	ServiceRoot string `env:"SERVICE_ROOT,required"`
}

func GetServiceConfig() *NimbusConf {
	var conf NimbusConf
	if err := envdecode.StrictDecode(&conf); err != nil {
		log.Fatalf("failed to decode: %s", err)
	}
	return &NimbusConf{
		ServiceRoot: conf.ServiceRoot,
	}
}

func (n *NimbusConf) GetServiceRoot() string {
	if n.ServiceRoot == "" {
		log.Fatalf("[ERROR] no service root set!")
		return ""
	}
	return n.ServiceRoot
}
