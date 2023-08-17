package db

import (
	"log"

	"github.com/joeshaw/envdecode"
)

// database configuration settings
type Conf struct {
	db dbConfig
}

type dbConfig struct {
	DBPath string `env:"DB_PATH,required"`
}

func DBConfig() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("[ERROR] failed to decode database config .env file: %s", err)
	}
	return &c
}
