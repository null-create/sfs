package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// used for run-time environment variable manipulation
type Env struct {
	fp  string
	env map[string]string
}

func NewE() *Env {
	if !HasDotEnv() {
		log.Fatal("no .env file present")
	}
	return &Env{
		fp:  filepath.Join(GetCwd(), ".env"),
		env: make(map[string]string),
	}
}

// read .env file and set as environment variables
func BuildEnv() error {
	if err := godotenv.Load(".env"); err != nil {
		return err
	}
	return nil
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

// make sure the environment variable matches whats defined in the .env file
func (e *Env) Validate(k string) error {
	env, err := godotenv.Read(".env")
	if err != nil {
		return err
	}
	if v, exists := env[k]; exists {
		val := os.Getenv(k)
		if val != v {
			msg := fmt.Sprintf("env mismatch. \n.env file (k=%v, v=%v) \nos.Getenv() (k=%s, v=%s)", k, v, k, val)
			return fmt.Errorf(msg)
		}
		return nil
	} else {
		return fmt.Errorf("%s key is not present in .env file", k)
	}
}

func (e *Env) Get(k string) (string, error) {
	env, err := godotenv.Unmarshal(e.fp)
	if err != nil {
		return "", err
	}
	if v, exists := env[k]; exists {
		val := os.Getenv(k) // make sure this is right
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

// update an environment variable and save to .env file
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
