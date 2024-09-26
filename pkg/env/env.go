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
	// general settings
	"ADMIN_MODE":        "false",
	"BUFFERED_EVENTS":   "true",
	"EVENT_BUFFER_SIZE": "2",
	"JWT_SECRET":        "",
	"NEW_SERVICE":       "true",

	// client settings
	"CLIENT_ADDRESS":      "localhost:9090",
	"CLIENT_BACKUP_DIR":   "",
	"CLIENT_EMAIL":        "",
	"CLIENT_ID":           "",
	"CLIENT_LOCAL_BACKUP": "false",
	"CLIENT_LOG_DIR":      "",
	"CLIENT_NAME":         "",
	"CLIENT_NEW_SERVICE":  "true",
	"CLIENT_PASSWORD":     "",
	"CLIENT_PORT":         "9090",
	"CLIENT_PROFILE_PIC":  "",
	"CLIENT_ROOT":         "",
	"CLIENT_TESTING":      "",
	"CLIENT_USERNAME":     "",

	// server settings
	"SERVER_ADDR":          "localhost:9191",
	"SERVER_ADMIN":         "admin",
	"SERVER_ADMIN_KEY":     "",
	"SERVER_HOST":          "",
	"SERVER_LOG_DIR":       "",
	"SERVER_PORT":          "9191",
	"SERVER_TIMEOUT_IDLE":  "900s",
	"SERVER_TIMEOUT_READ":  "5s",
	"SERVER_TIMEOUT_WRITE": "10s",

	// service settings
	"SERVICE_ENV":       "",
	"SERVICE_LOG_DIR":   "",
	"SERVICE_ROOT":      "",
	"SERVICE_TEST_ROOT": "",
}

// new env object.
// checks the current directory for the .env file by default.
func NewE() *Env {
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

func hasEnvFile() bool {
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

// check for the presence of a .env file in the given directory
func HasEnvFile(dirpath string) bool {
	entries, err := os.ReadDir(dirpath)
	if err != nil {
		log.Fatalf("failed to read directory entires: %v", err)
	}
	var found bool
	for _, e := range entries {
		if e.Name() == ".env" {
			found = true
		}
	}
	return found
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
			msg := fmt.Sprintf("env mismatch.\n.env file (k=%v, v=%v)\nos.Getenv() (k=%s, v=%s)", k, v, k, val)
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
		fmt.Printf("env var %v does not exist", k)
	}
	return nil
}

// Clears the environment configurations for both client and the server,
// then sets as a new .env file. use with caution! creates a backup copy
// after clearning, but the backup file is called 'env-backup.env', and will
// not be read by the service when started again after clearing.
func (e *Env) Clear() error {
	if err := Copy(".env", "env-backup.env"); err != nil {
		return err
	}
	e.env = BaseEnv
	if err := godotenv.Write(e.env, ".env"); err != nil {
		return err
	}
	for k, v := range e.env {
		if err := os.Setenv(k, v); err != nil {
			return err
		}
	}
	return nil
}

func (e *Env) List() {
	for k, v := range e.env {
		fmt.Printf("%v: %v\n", k, v)
	}
}
