package cmd

import (
	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
)

/*
Command to check whether there are any client-side items that aren't
registered with the server. If there are, the user will be prompted and if they
decide to register, then all the items will be registered.
*/

var (
	refreshCmd = &cobra.Command{
		Use:   "refresh",
		Short: "find and register any client-side items that aren't registered with the server",
		Run:   refresh,
	}
)

func init() {
	rootCmd.AddCommand(refreshCmd)
}

func refresh(cmd *cobra.Command, args []string) {
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(err)
		return
	}
	if err := c.Refresh(); err != nil {
		showerr(err)
	}
}
