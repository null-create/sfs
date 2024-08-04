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

func setDefaults(newEnv map[string]string) map[string]string {
	newEnv["ADMIN_MODE"] = "false"
	newEnv["BUFFERED_EVENTS"] = "true"
	newEnv["EVENT_BUFFER_SIZE"] = "2"
	newEnv["JWT_SECRET"] = auth.GenSecret(64)
	newEnv["NEW_SERVICE"] = "true"

	newEnv["CLIENT_ROOT"] = filepath.Join(Root, "pkg", "client", "run")
	newEnv["CLIENT_ID"] = auth.NewUUID()
	newEnv["CLIENT_ADDRESS"] = client.EndpointRoot + ":" + "8080"
	newEnv["CLIENT_LOCAL_BACKUP"] = "true"
	newEnv["CLIENT_AUTO_SYNC"] = "true"
	newEnv["CLIENT_BACKUP_DIR"] = filepath.Join(Root, "pkg", "client", "run", "backups")
	newEnv["CLIENT_LOG_DIR"] = filepath.Join(Root, "pkg", "client", "logs")
	newEnv["CLIENT_PORT"] = "8080"
	newEnv["CLIENT_NEW_SERVICE"] = "true"
	newEnv["CLIENT_TESTING"] = filepath.Join(Root, "pkg", "client", "testing")

	newEnv["SERVER_ADDR"] = client.EndpointRoot + ":" + "8080"
	newEnv["SERVER_ADMIN"] = "admin"
	newEnv["SERVER_ADMIN_KEY"] = auth.GenSecret(64)
	newEnv["SERVER_LOG_DIR"] = filepath.Join(Root, "pkg", "server", "logs")
	newEnv["SERVER_PORT"] = "8080"
	newEnv["SERVER_TIMEOUT_IDLE"] = "900s"
	newEnv["SERVER_TIMEOUT_READ"] = "5s"
	newEnv["SERVER_TIMEOUT_WRITE"] = "10s"

	newEnv["SERVICE_LOG_DIR"] = filepath.Join(Root, "pkg", "service", "logs")
	newEnv["SERVICE_ROOT"] = filepath.Join(Root, "pkg", "server", "run")
	newEnv["SERVICE_TEST_ROOT"] = filepath.Join(Root, "pkg", "server", "testing")
	return newEnv
}

// set configs and create .env files for each package
// populate baseEnv with default values for all fields
// get users input for client name, username, email
func setUpEnv() error {
	// set default values
	newEnv := setDefaults(env.BaseEnv)

	// manual settings set by the user
	for setting, value := range newEnv {
		if strings.Contains(setting, "CLIENT") && value == "" {
			var value string
			if setting == "CLIENT_NAME" {
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
	}

	// add .env file to root
	if err := env.NewEnvFile(filepath.Join(Root, ".env"), newEnv); err != nil {
		return err
	}
	// write out .env files to each package since they each need a copy
	// to execute their respective tests
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
