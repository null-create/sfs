package pkg

import (
	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/client"
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
func NewClient() (*client.Client, error) {
	cfg := client.ClientConfig()
	user := auth.NewUser(cfg.User, cfg.UserAlias, cfg.Email, cfg.Root, cfg.IsAdmin)
	return client.NewClient(user)
}

// intialize a new server side service instance
func NewServer() (*server.Service, error) {
	svc, err := server.Init(true, false)
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func SetUpEnv() error {
	// get users configurations via terminal for the first time set up

	// set configs and create .env files for each package

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
