package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/client"
	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/server"

	"github.com/spf13/cobra"
)

// command to execute a first time set up

var (
	setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "First time set up.",
		Long: `
First time set up. Creates a new client and server, and 
retrieves some basic information about the user to establish
client side services.
		`,
		Run: runSetupCmd,
	}
	// technique from:
	// https://stackoverflow.com/questions/31873396/is-it-possible-to-get-the-current-root-of-package-structure-as-a-string-in-golan
	_, b, _, _ = runtime.Caller(0)
	Root       = filepath.Join(filepath.Dir(b), "../..")
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetupCmd(cmd *cobra.Command, args []string) {
	// check for whether there are CLIENT configurations already.
	// we need to check for the presence of at least an .env file
	// at the root level, then parse it for CLIENT_ prefixes and
	// see if there are any values associated with those keys
	// if not, then we run the first time setup steps
	if env.CheckForDotEnv(Root) {
		// we only have an .env file if we've set up our configurations,
		// so we don't need to run first time setup again.
		cmdLogger.Info(".env file found. Setup has already been ran.")
		return
	}
	// set up environment configurations
	if err := setUpEnv(); err != nil {
		showerr(err)
		return
	}
	// set up server side service
	if err := newService(); err != nil {
		showerr(err)
		return
	}
	// set up client side service
	if err := newClient(); err != nil {
		showerr(err)
		return
	}
}

// set configs and create .env files for each package
// populate baseEnv with default values for all fields
// get users input for client name, username, email
func setUpEnv() error {
	newEnv := env.BaseEnv
	for setting, value := range newEnv {
		// client settings
		if strings.Contains(setting, "CLIENT") && value == "" {
			if setting == "CLIENT_ROOT" {
				newEnv[setting] = filepath.Join(Root, "pkg", "client", "run")
			} else if setting == "CLIENT_TESTING" {
				newEnv[setting] = filepath.Join(Root, "pkg", "client", "testing")
			} else if setting == "CLIENT_ID" {
				newEnv[setting] = auth.NewUUID()
			} else { // get the users inputs for specific settings
				var value string
				if setting == "CLIENT" {
					fmt.Print("Enter your name: ")
				} else if setting == "CLIENT_USERNAME" {
					fmt.Print("Enter your username: ")
				} else if setting == "CLIENT_EMAIL" {
					fmt.Print("Enter your email: ")
				} else if setting == "CLIENT_PASSWORD" {
					fmt.Print("Enter your password: ")
				}
				fmt.Scanln(&value)
				newEnv[setting] = value
			}
			// server and service settings
		} else if setting == "SERVER_ADMIN_KEY" {
			newEnv[setting] = auth.GenSecret(64)
		} else if setting == "SERVICE_ROOT" {
			newEnv[setting] = filepath.Join(Root, "pkg", "server", "run")
		} else if setting == "SERVICE_TEST_ROOT" {
			newEnv[setting] = filepath.Join(Root, "pkg", "server", "testing")
		} else if setting == "JWT_SECRET" {
			newEnv[setting] = auth.GenSecret(64)
		}
	}
	// Write out .env files to each package
	entries, err := os.ReadDir(filepath.Join(Root, "pkg"))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := env.NewEnvFile(filepath.Join(Root, "pkg", entry.Name(), ".env"), newEnv); err != nil {
			return err
		}
	}
	return nil
}

// create a new client service useing the .env file configurations
func newClient() error {
	_, err := client.Init(true)
	if err != nil {
		return err
	}
	return nil
}

func newService() error {
	_, err := server.Init(true, false)
	if err != nil {
		return err
	}
	return nil
}
