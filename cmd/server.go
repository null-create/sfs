package cmd

import (
	"os"

	"github.com/sfs/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	new   bool           // whether we should create a new sfs server
	start bool           // whether we should start the server
	stop  bool           // whether we should stop the server
	sig   chan os.Signal // shutdown channel for the server

	svcCfg = server.ServiceConfig() // server-side service configurations

	// server command
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "SFS Server Commands",
		RunE:  runCmd,
	}
)

func init() {
	serverCmd.PersistentFlags().BoolVar(&start, "start", false, "start the sfs server. stop with ctrl-c.")
	serverCmd.PersistentFlags().BoolVar(&stop, "stop", false, "stop the sfs server")
	serverCmd.PersistentFlags().BoolVarP(&new, "new", "n", false, "create a new sfs server side service instance")

	viper.BindPFlag("start", serverCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("stop", serverCmd.PersistentFlags().Lookup("stop"))
	viper.BindPFlag("new", serverCmd.PersistentFlags().Lookup("new"))

	rootCmd.AddCommand(serverCmd)
}

func runCmd(cmd *cobra.Command, args []string) error {
	newFlag, _ := cmd.Flags().GetBool("new")
	startFlag, _ := cmd.Flags().GetBool("start")
	stopFlag, _ := cmd.Flags().GetBool("stop")

	switch {
	case newFlag:
		if err := newService(); err != nil {
			return err
		}
	case startFlag:
		if err := startServer(); err != nil {
			return err
		}
	case stopFlag:
		sig <- os.Interrupt
	}
	return nil
}

func startServer() error {
	svr := server.NewServer()
	svr.Run()
	return nil
}

func newService() error {
	_, err := server.Init(svcCfg.NewService, svcCfg.IsAdmin)
	if err != nil {
		return err
	}
	return nil
}
