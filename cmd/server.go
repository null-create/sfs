package cmd

import (
	"fmt"

	"github.com/sfs/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	svcCfg = server.ServiceConfig() // server-side service configurations

	// server command
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "SFS Server Commands",
		Run:   RunServerCmd,
	}
)

func init() {
	flags := FlagPole{}
	serverCmd.PersistentFlags().BoolVarP(&flags.start, "start", "s", false, "start the sfs server. stop with ctrl-c.")
	serverCmd.PersistentFlags().BoolVarP(&flags.new, "new", "n", false, "create a new sfs server side service instance")

	viper.BindPFlag("start", serverCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("new", serverCmd.PersistentFlags().Lookup("new"))

	rootCmd.AddCommand(serverCmd)
}

func RunServerCmd(cmd *cobra.Command, args []string) {
	new, _ := cmd.Flags().GetBool("new")
	start, _ := cmd.Flags().GetBool("start")
	switch {
	case new:
		if err := newService(); err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
			return
		}
	case start:
		svr := server.NewServer()
		svr.Run()
	}
}
