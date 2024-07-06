package cmd

import (
	"fmt"
	"log"

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
	confCmd.PersistentFlags().StringVar(&flags.get, "get", "", "get client service configurations")
	confCmd.PersistentFlags().StringVar(&flags.set, "set", "", "set client service configurations")
	confCmd.PersistentFlags().BoolVar(&flags.show, "show", false, "show client service configurations")
	confCmd.PersistentFlags().StringVarP(&flags.value, "value", "v", "", "config setting value")

	viper.BindPFlag("get", confCmd.PersistentFlags().Lookup("get"))
	viper.BindPFlag("set", confCmd.PersistentFlags().Lookup("set"))
	viper.BindPFlag("show", confCmd.PersistentFlags().Lookup("show"))
	viper.BindPFlag("value", confCmd.PersistentFlags().Lookup("value"))

	// configurations are a sub-command of the root command
	rootCmd.AddCommand(confCmd)
}

func showSetting(setting string, value string) {
	fmt.Printf("\n%s = %s\n", setting, value)
}

func runConfCmd(cmd *cobra.Command, args []string) {
	settingToGet, _ := cmd.Flags().GetString("get")
	setting, _ := cmd.Flags().GetString("set")
	show, _ := cmd.Flags().GetBool("show")
	value, _ := cmd.Flags().GetString("value")

	switch {
	case show:
		envCfgs.List()
	case settingToGet != "":
		val, err := envCfgs.Get(settingToGet)
		if err != nil {
			showerr(err)
			return
		}
		showSetting(settingToGet, val)
	case setting != "":
		if value == "" {
			fmt.Printf("no value supplied for setting %s", setting)
			return
		}
		if err := envCfgs.Set(setting, value); err != nil {
			showerr(err)
			return
		}
		log.Printf("setting %s changed to %s", setting, value)
	}
}
