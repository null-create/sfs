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
	toRemove string

	RemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove files or directories from the SFS filesystem",
		Run:   RunRemoveCmd,
	}
)

func init() {
	RemoveCmd.PersistentFlags().StringVar(&toRemove, "remove", "", "Remove files or directories from the SFS filesystem")

	viper.BindPFlag("remove", RemoveCmd.Flags().Lookup("remove"))

	clientCmd.AddCommand(RemoveCmd)
}

func RunRemoveCmd(cmd *cobra.Command, args []string) {
	remove, _ := cmd.Flags().GetString("remove")
	if remove == "" {
		showerr(fmt.Errorf("no path specified"))
	}
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(err)
	}
	if err := c.RemoveItem(remove); err != nil {
		showerr(err)
	}
}
