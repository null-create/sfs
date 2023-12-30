package pkg

/*
File for creating all necessary .env files for all packages.
Should be run the first time Simple File Sync is installed.

Also will contain a "repair" function that will check for any missing
files and re-create them if necessary. This may involve pulling
from the sfs github repo and/or using some sort of file validation
system. Will need to investigate.
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

// TODO: create/generate some sort of exteranal file listing of all
// necessary SFS files/packages? Will/can go pull this info from github?
func Repair() error { return nil }
