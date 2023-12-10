package pkg

/*
File for creating all necessary .env files for all packages
Should be run the first time Simple File Sync is installed.

Also will contain a "repair" function that will check for any missing
files and re-create them if necessary
*/

// base environment which will need to be customized
// to the user's preferences
var baseEnv = map[string]string{
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

// creates .env files for all packages
func SetUpSfs() error {
	// cwd, _ := os.Getwd()

	// populate baseEnv with default values for all fields

	// create an .env file for all packages except bin, env, models,

	return nil
}

// names of all required internal packages
var packages = []string{
	"auth", "bin", "client", "db", "env", "models",
	"monitor", "network", "server", "service", "transfer",
}

func Repair() error { return nil }
