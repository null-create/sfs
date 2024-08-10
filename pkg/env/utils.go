package env

import (
	"io"
	"log"
	"os"
)

func GetCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("could not get current working directory: %v", err)
	}
	return dir
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
