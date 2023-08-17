package network

import (
	"log"
	"os"
	"path/filepath"
)

func ProfileDirPath() string {
	return filepath.Join(GetCwd(), "profile")
}

func GetCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("[ERROR] unable to get current working directory %v", err)
	}
	return dir
}

func GetHostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("[ERROR] unable to get hostname \n%v\n ", err)
		return ""
	}
	return hostname
}
