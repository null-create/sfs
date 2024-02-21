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
	id string

	RemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove files or directories from the SFS filesystem",
		Run:   RunRemoveCmd,
	}
)

func init() {
	RemoveCmd.PersistentFlags().StringVar(&name, "name", "", "Remove files or directories from the SFS filesystem using their name")
	RemoveCmd.PersistentFlags().StringVar(&id, "id", "", "Remove file or directory from the SFS filesystem using their ID")

	viper.BindPFlag("remove", RemoveCmd.Flags().Lookup("remove"))

	clientCmd.AddCommand(RemoveCmd)
}

func RunRemoveCmd(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	id, _ := cmd.Flags().GetString("id")
	if name == "" || id == "" {
		showerr(fmt.Errorf("need either name or id to remove. name=%v id=%s", name, id))
		return
	}
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(err)
	}
	if err := c.RemoveItem(name); err != nil {
		showerr(err)
	}
}
