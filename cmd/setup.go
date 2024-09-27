package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/client"
	"github.com/sfs/pkg/env"
	"github.com/sfs/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// command to execute a first time set up

var (
	setupCmd = &cobra.Command{
		Use:   "setup",
		Short: "First time set up.",
		Long: `
First time set up after building. 

Creates necessary .env files based on the specifications
defined in pkg/configs/configs.yaml

Use the -a flag to automatically generate the .env files.
Use the -d flag to specify where the SFS application binary should be located.

CLIENT_NAME and CLIENT_USERNAME will be randomly generated.
		`,
		Run: runSetupCmd,
	}
	auto bool
)

func init() {
	setupCmd.Flags().BoolVarP(&auto, "auto", "a", false, "Whether to automate baseline environment configs (defaults to false)")

	viper.BindPFlag("auto", setupCmd.Flags().Lookup("auto"))

	rootCmd.AddCommand(setupCmd)
}

func runSetupCmd(cmd *cobra.Command, args []string) {
	auto, _ := cmd.Flags().GetBool("auto")
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if !env.HasEnvFile(cwd) {
		if err := setUpEnv(auto, cwd); err != nil {
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
	} else {
		cmdLogger.Info("environment configuration files found. skipping setup.")
		return
	}
}

func setDefaults(newEnv map[string]string, root string) map[string]string {
	newEnv["ADMIN_MODE"] = "false"
	newEnv["BUFFERED_EVENTS"] = "true"
	newEnv["EVENT_BUFFER_SIZE"] = "2"
	newEnv["JWT_SECRET"] = auth.GenSecret(64)
	newEnv["NEW_SERVICE"] = "true"
	newEnv["CLIENT_ADDRESS"] = client.EndpointRoot + ":" + "9090"
	newEnv["CLIENT_BACKUP_DIR"] = filepath.Join(root, "pkg", "client", "run", "backups")
	newEnv["CLIENT_ROOT"] = filepath.Join(root, "pkg", "client", "run")
	newEnv["CLIENT_ID"] = auth.NewUUID()
	newEnv["CLIENT_NEW_SERVICE"] = "true"
	newEnv["CLIENT_LOCAL_BACKUP"] = "false"
	newEnv["CLIENT_LOG_DIR"] = filepath.Join(root, "pkg", "client", "logs")
	newEnv["CLIENT_PASSWORD"] = auth.GenSecret(64)
	newEnv["CLIENT_PORT"] = "9090"
	newEnv["CLIENT_TESTING"] = filepath.Join(root, "pkg", "client", "testing")
	newEnv["SERVER_ADDR"] = client.EndpointRoot + ":" + "9191"
	newEnv["SERVER_ADMIN"] = "admin"
	newEnv["SERVER_ADMIN_KEY"] = auth.GenSecret(64)
	newEnv["SERVER_LOG_DIR"] = filepath.Join(root, "pkg", "server", "logs")
	newEnv["SERVER_PORT"] = "9191"
	newEnv["SERVER_TIMEOUT_IDLE"] = "900s"
	newEnv["SERVER_TIMEOUT_READ"] = "5s"
	newEnv["SERVER_TIMEOUT_WRITE"] = "10s"
	newEnv["SERVICE_ENV"] = filepath.Join(root, "pkg", "env", ".env")
	newEnv["SERVICE_LOG_DIR"] = filepath.Join(root, "pkg", "service", "logs")
	newEnv["SERVICE_ROOT"] = filepath.Join(root, "pkg", "server", "run")
	newEnv["SERVICE_TEST_ROOT"] = filepath.Join(root, "pkg", "server", "testing")
	return newEnv
}

// set configs and create .env files for each package
// populate baseEnv with default values for all fields
// get users input for client name, username, email
func setUpEnv(auto bool, root string) error {
	// new baseline environment configurations
	newEnv := setDefaults(env.BaseEnv, root)

	// get users input for client name, username, email, or
	// generate automatic defaults depending on the auto param
	for setting, value := range newEnv {
		if strings.Contains(setting, "CLIENT") && value == "" {
			if !auto {
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
				_, err := fmt.Scanln(&value)
				if err != nil {
					return err
				}
				newEnv[setting] = value
			} else {
				if setting == "CLIENT_NAME" {
					newEnv[setting] = "someone"
				} else if setting == "CLIENT_USERNAME" {
					newEnv[setting] = "a_username"
				} else if setting == "CLIENT_EMAIL" {
					newEnv[setting] = "anemail@example.com"
				}
			}
		}
	}

	// add .env file to root
	if err := env.NewEnvFile(filepath.Join(root, ".env"), newEnv); err != nil {
		return err
	}
	// write out .env files to each package since they each need a copy
	// to execute their respective tests
	entries, err := os.ReadDir(filepath.Join(root, "pkg"))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if err := env.NewEnvFile(filepath.Join(root, "pkg", entry.Name(), ".env"), newEnv); err != nil {
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
