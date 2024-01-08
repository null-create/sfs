package cmd

import (
	"github.com/sfs/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	new      bool      // whether we should create a new sfs server
	start    bool      // whether we should start the server
	stop     bool      // whether we should stop the server
	shutDown chan bool // shutdown channel for the server

	scfg = server.ServiceConfig() // server-side service configurations

	// server command
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "SFS Server Commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			newFlag, _ := cmd.Flags().GetBool("new")
			startFlag, _ := cmd.Flags().GetBool("start")
			stopFlag, _ := cmd.Flags().GetBool("stop")
			switch {
			case newFlag:
				return newServer()
			case startFlag:
				return startServer()
			case stopFlag:
				shutDown <- true
			}
			return nil
		},
	}
)

// TODO: make this a non-blocking persistent background process.
// need to shut down via command line, not ctrl-c
func startServer() error {
	svr := server.NewServer()
	svr.Start(shutDown)
	return nil
}

func newServer() error {
	_, err := server.Init(scfg.NewService, scfg.IsAdmin)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	serverCmd.PersistentFlags().BoolVar(&start, "start", false, "start the sfs server")
	serverCmd.PersistentFlags().BoolVar(&stop, "stop", false, "stop the sfs server")
	serverCmd.PersistentFlags().BoolVarP(&new, "new", "n", false, "create a new sfs server side service instance")

	viper.BindPFlag("start", serverCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("stop", serverCmd.PersistentFlags().Lookup("stop"))
	viper.BindPFlag("new", serverCmd.PersistentFlags().Lookup("new"))

	rootCmd.AddCommand(serverCmd)
}
