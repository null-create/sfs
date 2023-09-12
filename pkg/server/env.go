package server

import (
	"log"

	"github.com/joho/godotenv"
)

// read .env file and set as environment variables
func BuildEnv() error {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("could not load .env file: %v", err)
	}
	return nil
}
