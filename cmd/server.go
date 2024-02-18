package cmd

import (
	"github.com/sfs/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	new   bool // whether we should create a new sfs server
	start bool // whether we should start the server

	svcCfg = server.ServiceConfig() // server-side service configurations

	// server command
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "SFS Server Commands",
		Run:   RunServerCmd,
	}
)

func init() {
	serverCmd.PersistentFlags().BoolVarP(&start, "start", "s", false, "start the sfs server. stop with ctrl-c.")
	serverCmd.PersistentFlags().BoolVarP(&new, "new", "n", false, "create a new sfs server side service instance")

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
			showerr(err)
			return
		}
	case start:
		startServer()
	}
}

func startServer() {
	svr := server.NewServer()
	svr.Run()
}

func newService() error {
	_, err := server.Init(svcCfg.NewService, svcCfg.IsAdmin)
	if err != nil {
		return err
	}
	return nil
}
