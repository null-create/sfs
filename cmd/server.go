package cmd

import (
	"fmt"

	"github.com/sfs/pkg/server"

	"github.com/spf13/cobra"
)

var (
	svr      *server.Server // pointer to server instance. used for run time info.
	shutDown chan bool      // shutdown channel for the server

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
			sfs server stats   - get the run time stats for the server
		`,
	}

	// ------ child commands

	newSvrCmd = &cobra.Command{
		Use:   "new",
		Short: "create a new sfs server side service instance",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
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
}
