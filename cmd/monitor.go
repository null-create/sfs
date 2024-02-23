package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Command for managing the monitoring service
*/

var (
	monCmd = &cobra.Command{
		Use:   "monitor",
		Short: "Command for managing the SFS monitoring service",
		Run:   RunMonCmd,
	}
)

func init() {
	flags := FlagPole{}
	monCmd.PersistentFlags().StringVar(&flags.ignore, "ignore", "", "ignore a file or directory")

	viper.BindPFlag("ignore", monCmd.PersistentFlags().Lookup("ignore"))

	clientCmd.AddCommand(monCmd)
}

func RunMonCmd(cmd *cobra.Command, args []string) {

}
