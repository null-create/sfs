package cmd

import (
	"github.com/sfs/pkg/client"
	"github.com/sfs/pkg/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	svr *server.Server // pointer to server instance. used for reading run time info.

	new      bool      // whether we should create a new sfs server
	start    bool      // whether we should start the server
	stop     bool      // whether we should stop the server
	shutDown chan bool // shutdown channel for the server

	conf = client.ClientConfig() // client service configurations

	// ------ server command

	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "SFS Server Commands",
		Long: `
		SFS Server command. 
	
		Use the flags associated with this command to configure and manage the SFS server.
	
		Examples: 
			sfs server --new     - create a new sfs server side service instance.
			sfs server --start   - start the SFS server
			sfs server --stop    - shutdown the SFS server
			sfs server --stats   - get the run time stats for the server
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			startFlag, _ := cmd.Flags().GetBool("start")
			stopFlag, _ := cmd.Flags().GetBool("stop")
			newFlag, _ := cmd.Flags().GetBool("new")

			switch {
			case startFlag:
				svr = server.NewServer()
				go func() {
					svr.Start(shutDown)
				}()
			case stopFlag:
				shutDown <- true
			case newFlag:
				_, err := server.Init(conf.NewService, conf.IsAdmin)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
)

func init() {
	serverCmd.PersistentFlags().BoolVar(&start, "start", false, "start the sfs server")
	serverCmd.PersistentFlags().BoolVar(&stop, "stop", false, "stop the sfs server")
	serverCmd.PersistentFlags().BoolVar(&new, "new", false, "create a new sfs server side service instance")

	viper.BindPFlag("start", serverCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("stop", serverCmd.PersistentFlags().Lookup("stop"))
	viper.BindPFlag("new", serverCmd.PersistentFlags().Lookup("new"))

	rootCmd.AddCommand(serverCmd)
}
