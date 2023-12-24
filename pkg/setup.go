package pkg

/*
File for creating all necessary .env files for all packages
Should be run the first time Simple File Sync is installed.

Also will contain a "repair" function that will check for any missing
files and re-create them if necessary
*/

// creates .env files for all packages
func SetUpSfs() error {
	// cwd, _ := os.Getwd()

	// populate baseEnv with default values for all fields

	// create an .env file for all packages except bin, env, models,

	return nil
}

// names of all required internal packages
var Packages = []string{
	"auth", "bin", "client", "db", "env", "models",
	"monitor", "network", "server", "service", "transfer",
}

func Repair() error { return nil }
