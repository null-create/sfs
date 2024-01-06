package cmd

import (
	"fmt"

	"github.com/sfs/pkg/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// flag variables
	start bool
	stop  bool
	stats bool
	new   bool

	svr *server.Server // pointer to server instance. used for run time info.

	shutDown chan bool // shutdown channel for the server

	// ------ server command
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "sfs server commands",
		Long: `
		SFS Server command. 
	
		Use the flags associated with this command to configure and manage the SFS server.
	
		Examples: 
			sfs server new     - create a new sfs server side service instance.
			sfs server start   - start the SFS server
			sfs server stop    - shutdown the SFS server
		`,
	}

	// ------ child commands

	newSvrCmd = &cobra.Command{
		Use:   "new",
		Short: "create a new sfs server side service instance",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: handle admin param as a flag
			_, err := server.Init(true, false)
			if err != nil {
				return err
			}
			return nil
		},
	}

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "start the SFS server",
		RunE: func(cmd *cobra.Command, args []string) error {
			svr = server.NewServer()
			go func() {
				svr.Start(shutDown)
			}()
			return nil
		},
	}

	stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "stop the SFS server",
		Run: func(cmd *cobra.Command, args []string) {
			shutDown <- true
		},
	}

	statsCmd = &cobra.Command{
		Use:   "stats",
		Short: "stats the SFS server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("server run time %s", svr.RunTime())
		},
	}
)

func init() {
	// add command and child commands to root command
	serverCmd.AddCommand(newSvrCmd)
	serverCmd.AddCommand(startCmd)
	serverCmd.AddCommand(stopCmd)
	serverCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(serverCmd)

	// add flags
	serverCmd.PersistentFlags().BoolVarP(&new, "new", "n", true, "set up a new server side sfs service instance")
	serverCmd.PersistentFlags().BoolVar(&start, "start", false, "flag to start the sfs server")
	serverCmd.PersistentFlags().BoolVar(&stop, "start", false, "flag to stop the sfs server")
	serverCmd.PersistentFlags().BoolVar(&stats, "stats", false, "flag to display run time health of the server")

	viper.BindPFlag("new", serverCmd.Flags().Lookup("new"))
	viper.BindPFlag("start", serverCmd.Flags().Lookup("start"))
	viper.BindPFlag("stop", serverCmd.Flags().Lookup("stop"))
	viper.BindPFlag("stats", serverCmd.Flags().Lookup("stats"))
}
