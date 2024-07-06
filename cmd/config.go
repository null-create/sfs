package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

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
	configs = client.GetClientConfigs()

	confCmd = &cobra.Command{
		Use:   "conf",
		Short: "Get, set, and view client service configurations",
		Run:   runConfCmd,
	}
)

func init() {
	flags := FlagPole{}
	confCmd.PersistentFlags().BoolVar(&flags.get, "get", false, "get client service configurations")
	confCmd.PersistentFlags().BoolVar(&flags.set, "set", false, "set client service configurations")
	confCmd.PersistentFlags().BoolVar(&flags.show, "show", true, "show client service configurations")
	confCmd.PersistentFlags().StringVarP(&flags.key, "setting", "s", "", "config setting name")
	confCmd.PersistentFlags().StringVarP(&flags.value, "value", "v", "", "config setting value")

	viper.BindPFlag("get", confCmd.PersistentFlags().Lookup("get"))
	viper.BindPFlag("set", confCmd.PersistentFlags().Lookup("set"))
	viper.BindPFlag("show", confCmd.PersistentFlags().Lookup("show"))
	viper.BindPFlag("setting", confCmd.PersistentFlags().Lookup("setting"))
	viper.BindPFlag("value", confCmd.PersistentFlags().Lookup("value"))

	// configurations are a sub-command of the client command
	clientCmd.AddCommand(confCmd)
}

func settingHasValue(cmd *cobra.Command, setting string) bool {
	val, _ := cmd.Flags().GetString(setting)
	return val != ""
}

func showSetting(setting string, value string) {
	fmt.Printf("%s = %s\n", setting, value)
}

func showClientSettings() {
	var output string
	env := os.Environ()
	for _, e := range env {
		if strings.Contains(e, "CLIENT") {
			output += e + "\n"
		}
	}
	fmt.Print(output)
}

func runConfCmd(cmd *cobra.Command, args []string) {
	get, _ := cmd.Flags().GetBool("get")
	set, _ := cmd.Flags().GetBool("set")
	show, _ := cmd.Flags().GetBool("show")
	setting, _ := cmd.Flags().GetString("setting")
	value, _ := cmd.Flags().GetString("value")

	switch {
	case show:
		envCfgs.List()
	case get:
		if !settingHasValue(cmd, setting) {
			showerr(fmt.Errorf("no value supplied for setting %s", setting))
			return
		}
		val, err := envCfgs.Get(setting)
		if err != nil {
			showerr(err)
			return
		}
		showSetting(setting, val)
	case set:
		if !settingHasValue(cmd, setting) {
			showerr(fmt.Errorf("no value supplied for setting %s", setting))
			return
		}
		if err := envCfgs.Set(setting, value); err != nil {
			showerr(err)
			return
		}
		log.Printf("setting %s changed to %s", setting, value)
	}
}
