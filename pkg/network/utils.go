package network

import (
	"io"
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
	}
	return hostname
}

// copy a file
func Copy(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}
	return nil
}
