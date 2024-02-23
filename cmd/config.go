package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Command for setting and getting client configurations
*/

var (
	configs = client.ClientConfig()

	confCmd = &cobra.Command{
		Use:   "conf",
		Short: "set and get client service configurations",
		Run:   RunConfCmd,
	}
)

func init() {
	flags := FlagPole{}
	confCmd.PersistentFlags().BoolVar(&flags.get, "get", false, "get client service configurations")
	confCmd.PersistentFlags().BoolVar(&flags.set, "set", false, "set client service configurations")

	viper.BindPFlag("get", confCmd.PersistentFlags().Lookup("get"))
	viper.BindPFlag("set", confCmd.PersistentFlags().Lookup("set"))

	// configurations are a sub-command of the client command
	clientCmd.AddCommand(confCmd)
}

func RunConfCmd(cmd *cobra.Command, args []string) {
	get, _ := cmd.Flags().GetBool("get")
	set, _ := cmd.Flags().GetBool("set")

	switch {
	case get:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		c.GetConfigs()
	case set:
		fmt.Print("not implemented yet.")
	}
}
