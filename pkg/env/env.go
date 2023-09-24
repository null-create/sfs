package env

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// used for run-time environment variable manipulation
type Env map[string]string

func NewE() *Env {
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
		return err
	}
	return nil
}

func (e *Env) validate(k string, env map[string]string) error {
	if v, exists := env[k]; exists {
		val := os.Getenv(k) // make sure this is right
		if val != v {
			msg := fmt.Sprintf("env mismatch. \n.env file (k=%v, v=%v) \nos.Getenv() (k=%s, v=%s)", k, v, k, val)
			return fmt.Errorf(msg)
		}
		return nil
	} else {
		return fmt.Errorf("%s not found", k)
	}
}

// make sure the environment variable matches whats defined in the .env file
func (e *Env) Validate(k string) error {
	env, err := godotenv.Read(".env")
	if err != nil {
		return err
	} else {
		return e.validate(k, env)
	}
}

func (e *Env) Get(k string) (string, error) {
	env, err := godotenv.Read(".env")
	if err != nil {
		return "", err
	}
	if err := e.validate(k, env); err != nil {
		return env[k], nil
	} else {
		return "", fmt.Errorf("%s not found: %v", k, err)
	}
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
	env, err := godotenv.Read(".env")
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
