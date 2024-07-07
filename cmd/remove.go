package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Command for removing files or directories from the SFS filesystem
*/

var (
	RemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove files or directories from the SFS filesystem",
		Long: `
Remove files or directories from the SFS filesystem. Attempts to remove
physical files and directories from both the client and the server. 

***Use with caution!***

Files get copied to the SFS recycle bin directory on the client side, and 
will be removed from their original location on the users machine.`,
		Run: removeCmd,
	}
)

func init() {
	flags := FlagPole{}
	RemoveCmd.PersistentFlags().StringVarP(
		&flags.path, "path", "p", "", "Delete files or directories using their absolute paths",
	)
	viper.BindPFlag("path", RemoveCmd.Flags().Lookup("path"))

	drvCmd.AddCommand(RemoveCmd)
}

func removeCmd(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("path")
	if path == "" {
		showerr(fmt.Errorf("path was not provided"))
		return
	}
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(fmt.Errorf("failed to initialize service: %v", err))
	}
	// move to recycle bin
	if err := c.RemoveItem(path); err != nil {
		showerr(fmt.Errorf("failed to remove item: %v", err))
	}
}
