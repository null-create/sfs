package env

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// used for run-time environment variable manipulation
type Env struct {
	env map[string]string // environment map
	loc string            // location of the .env file
}

// base environment which will need to be customized
// to the user's preferences
var BaseEnv = map[string]string{
	"ADMIN_MODE":           "false",
	"CLIENT":               "",
	"CLIENT_USERNAME":      "",
	"CLIENT_ADDRESS":       "",
	"CLIENT_EMAIL":         "",
	"CLIENT_ID":            "",
	"CLIENT_NEW_SERVICE":   "true",
	"CLIENT_PASSWORD":      "default",
	"CLIENT_PORT":          "8080",
	"CLIENT_ROOT":          "",
	"CLIENT_TESTING":       "",
	"JWT_SECRET":           "default",
	"NEW_SERVICE":          "true",
	"SERVER_ADDR":          "localhost:8080",
	"SERVER_ADMIN":         "admin",
	"SERVER_ADMIN_KEY":     "default",
	"SERVER_PORT":          "8080",
	"SERVER_TIMEOUT_IDLE":  "900s",
	"SERVER_TIMEOUT_READ":  "5s",
	"SERVER_TIMEOUT_WRITE": "10s",
	"SERVICE_ROOT":         "",
	"SERVICE_TEST_ROOT":    "",
}

// new env object.
// checks the current directory for the .env file by default.
func NewE() *Env {
	if !HasDotEnv() {
		log.Fatal("no .env file present")
	}
	env, err := godotenv.Read(".env")
	if err != nil {
		log.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	return &Env{
		env: env,
		loc: filepath.Join(cwd, ".env"),
	}
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

// create a new baseline .env file using the provided configurations.
// does not set the environment. use env.SetEnv(false) after a call to this if needed.
//
// saves in the designated directory.
func NewEnvFile(path string, env map[string]string) error {
	return godotenv.Write(env, filepath.Join(path, ".env"))
}

// read .env file and set as environment variables
func SetEnv(debug bool) {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}
	if debug {
		env := os.Environ()
		for i, e := range env {
			fmt.Printf("%d: %s\n", i+1, e)
		}
	}
}

func validate(k string, env map[string]string) error {
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
		return validate(k, env)
	}
}

func (e *Env) Get(k string) (string, error) {
	env, err := godotenv.Read(".env")
	if err != nil {
		return "", err
	}
	if err := validate(k, env); err == nil {
		return env[k], nil
	} else {
		return "", fmt.Errorf("%s not found: %v", k, err)
	}
}

func (e *Env) GetEnv() (map[string]string, error) {
	env, err := godotenv.Read(".env")
	if err != nil {
		return nil, err
	}
	// make sure this is accurate first
	for _, k := range env {
		if err := validate(k, env); err != nil {
			return nil, err
		}
	}
	return env, nil
}

func set(k, v string, env map[string]string) error {
	if err := godotenv.Write(env, ".env"); err != nil {
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
