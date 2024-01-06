package cmd

import "github.com/spf13/cobra"

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "client commands",
	Long: `
	SFS Client Commands. Use the flags associated with this command
	to manage the clients local filesystem and execute synchornization
	operations between the client and server.
	
	Examples:

	sfs client --start   - Start the SFS client
	sfs client --stop    - Stop the SFS client
	sfs client --list    - List root entries
	sfs client --sync    - Sync local files with the SFS server
	`,
	Run: func(c *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(clientCmd)
}
