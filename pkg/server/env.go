package server

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// used for run-time environment variable manipulation
type Env map[string]string

func NewE() *Env {
	// make sure the .env file exists before instantiating
	if !HasDotEnv() {
		log.Fatal("no .env file present")
	}
	return &Env{}
}

func HasDotEnv() bool {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working directory: %v", err)
	}
	entries, err := os.ReadDir(wd)
	if err != nil {
		log.Fatalf("failed to read directory entires: %v", err)
	}
	for _, e := range entries {
		if e.Name() == ".env" {
			return true
		}
	}
	return false
}

// read .env file and set as environment variables
func BuildEnv() error {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("could not load .env file: %v", err)
	}
	return nil
}

func (e *Env) Get(k string) (string, error) {
	env, err := godotenv.Unmarshal(".env")
	if err != nil {
		return "", err
	}
	if v, exists := env[k]; exists {
		// make sure this is right
		val := os.Getenv(k)
		if val != v {
			msg := fmt.Sprintf("env mismatch. \n.env file (k=%v, v=%v) \nos.Getenv() (k=%s, v=%s)", k, v, k, val)
			return "", fmt.Errorf(msg)
		}
		return v, nil
	}
	return "", fmt.Errorf("%s not found", k)
}

func set(k, v string, env map[string]string) error {
	_, err := godotenv.Marshal(env)
	if err != nil {
		return err
	}
	if err := os.Setenv(k, v); err != nil {
		return err
	}
	return nil
}

// update an environment variable, and save to .env file for later use
func (e *Env) Set(k, v string) error {
	env, err := godotenv.Unmarshal(".env")
	if err != nil {
		return err
	}
	if val, exists := env[k]; exists {
		if val != v {
			env[k] = v
			if err := set(k, v, env); err != nil {
				return err
			}
		}
	} else {
		log.Printf("env var %v does not exist", k)
	}
	return nil
}

// loads the .env file and an env: map[string]string
func (e *Env) Load() (map[string]string, error) {
	env, err := godotenv.Unmarshal(".env")
	if err != nil {
		return nil, err
	}
	return env, err
}
