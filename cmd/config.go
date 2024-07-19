package cmd

import (
	"fmt"
	"log"
	"strconv"

	"github.com/sfs/pkg/client"
	"github.com/sfs/pkg/env"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Command for setting and getting client configurations
*/

var (
	envCfgs = env.NewE()

	confCmd = &cobra.Command{
		Use:   "conf",
		Short: "Get, set, and view SFS service configurations",
		Run:   runConfCmd,
	}
)

func init() {
	flags := FlagPole{}
	confCmd.PersistentFlags().StringVarP(&flags.get, "get", "g", "", "get client service configurations")
	confCmd.PersistentFlags().StringVarP(&flags.setting, "set", "s", "", "set a client service configuration")
	confCmd.PersistentFlags().BoolVarP(&flags.list, "list", "l", false, "show client service configurations")
	confCmd.PersistentFlags().StringVarP(&flags.value, "value", "v", "", "config setting value")

	viper.BindPFlag("get", confCmd.PersistentFlags().Lookup("get"))
	viper.BindPFlag("set", confCmd.PersistentFlags().Lookup("set"))
	viper.BindPFlag("list", confCmd.PersistentFlags().Lookup("list"))
	viper.BindPFlag("value", confCmd.PersistentFlags().Lookup("value"))

	// configurations are a sub-command of the root command
	rootCmd.AddCommand(confCmd)
}

func showSetting(setting string, value string) {
	fmt.Printf("\n%s = %s\n", setting, value)
}

// see if local backup mode is enabled first
// if so, thent he client won't have files stored on the sfs server
func localBackupIsEnabled() bool {
	b, err := envCfgs.Get("CLIENT_LOCAL_BACKUP")
	if err != nil {
		showerr(err)
		return false
	}
	return b == "true"
}

func getConfigFlags(cmd *cobra.Command) *FlagPole {
	settingToGet, _ := cmd.Flags().GetString("get")
	settingToSet, _ := cmd.Flags().GetString("set")
	list, _ := cmd.Flags().GetBool("list")
	value, _ := cmd.Flags().GetString("value")

	return &FlagPole{
		setting: settingToSet,
		get:     settingToGet,
		list:    list,
		value:   value,
	}
}

func runConfCmd(cmd *cobra.Command, args []string) {
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(fmt.Errorf("failed to initialize service: %v", err))
	}

	f := getConfigFlags(cmd)

	switch {
	case f.list:
		envCfgs.List()
	// display a setting
	case f.get != "":
		val, err := envCfgs.Get(f.get)
		if err != nil {
			showerr(err)
			return
		}
		showSetting(f.get, val)
	// handle updates to various settings
	case f.setting != "":
		if f.value == "" {
			fmt.Printf("no value supplied for setting '%s'", f.setting)
			return
		}
		switch f.setting {
		case "CLIENT_LOCAL_BACKUP":
			val, err := strconv.ParseBool(f.value)
			if err != nil {
				showerr(err)
				return
			}
			c.SetLocalBackup(val)
		case "CLIENT_AUTO_SYNC":
			val, err := strconv.ParseBool(f.value)
			if err != nil {
				showerr(err)
				return
			}
			c.SetAutoSync(val)
		case "CLIENT_BACKUP_DIR":
			if err := c.UpdateBackupPath(f.value); err != nil {
				log.Fatal(err)
			}
		}
		// update environment configurations
		if err := envCfgs.Set(f.setting, f.value); err != nil {
			showerr(err)
			return
		}
		log.Printf("setting %s changed to %s", f.setting, f.value)
	}
}
