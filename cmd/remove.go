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
		Run:   RunRemoveCmd,
	}
)

func init() {
	flags := FlagPole{}
	RemoveCmd.PersistentFlags().StringVar(&flags.path, "path", "", "Remove files or directories from the SFS filesystem using their absolute paths")
	RemoveCmd.PersistentFlags().BoolVar(&flags.delete, "delete", false, "set to true to delete, false to just stop monitoring")

	viper.BindPFlag("path", RemoveCmd.Flags().Lookup("path"))
	viper.BindPFlag("delete", RemoveCmd.Flags().Lookup("delete"))

	clientCmd.AddCommand(RemoveCmd)
}

func RunRemoveCmd(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("path")
	delete, _ := cmd.Flags().GetBool("delete")
	if path == "" {
		showerr(fmt.Errorf("path was not provided"))
		return
	}
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(err)
	}
	if delete {
		if err := c.RemoveItem(path); err != nil {
			showerr(err)
		}
	} else {
		c.Monitor.StopWatching(path)
		// TODO: create an "ignore" list.
	}
}
