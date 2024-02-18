package cmd

import (
	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
)

/*
File for initiating sync operations with the server
*/

var (
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Initiates a sync operation between the client and the server",
		Run:   SyncCmd,
	}
)

func init() {
	clientCmd.AddCommand(syncCmd)
}

func SyncCmd(cmd *cobra.Command, args []string) {
	c, err := client.LoadClient(true)
	if err != nil {
		showerr(err)
	}
	if err := c.Sync(); err != nil {
		showerr(err)
	}
}
