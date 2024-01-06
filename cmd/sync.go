package cmd

import "github.com/spf13/cobra"

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "client sync commands",
	Long: `
	SFS Client side synchronization command. 
	
	Use the flags associated with this command to configure and execute
	Synchronization operations from the client to the server.
	`,
	Run: func(c *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
