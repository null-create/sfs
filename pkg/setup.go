package pkg

import (
	"path/filepath"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/client"
	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/server"
)

/*
File for retrieving users client and server configurations then creating
all necessary .env files for all packages with these configs.

Should be run the first time Simple File Sync is installed.

Also will contain a "repair" function that will check for any missing
files and re-create them if necessary. This may involve pulling
from the sfs github repo and/or using some sort of file validation
system. Will need to investigate.
*/

// create a new client service useing the .env file configurations
var newClient = func() (*client.Client, error) {
	cfg := client.ClientConfig()
	user := auth.NewUser(cfg.User, cfg.UserAlias, cfg.Email, cfg.Root, cfg.IsAdmin)
	return client.NewClient(user)
}

// intialize a new server side service instance
var newServer = func() (*server.Service, error) {
	svc, err := server.Init(true, false)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

var setUpEnv = func() error {
	// get users configurations via terminal for the first time set up

	// set configs and create .env files for each package
	cwd := client.GetWd()
	for _, pkg := range Packages {
		// TODO: change 2nd arg to configs env map
		if err := env.NewEnvFile(filepath.Join(cwd, pkg), nil); err != nil {
			return err
		}
	}
	return nil
}

// creates .env files for all packages
func FirstTimeSetup() error {
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
