package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Command to automatically add all items under a given directory (and subdirectories)
To the SFS service. Does not have to be the SFS root! Can be any directory
specified by the user.
*/

var (
	discoverCmd = &cobra.Command{
		Use:   "discover",
		Short: "Use to automatically discover and add all items under a given directory",
		Run:   RunDiscoverCmd,
	}
)

func init() {
	flags := FlagPole{}
	discoverCmd.Flags().StringVar(&flags.path, "path", "", "Path to the directory to run discover on")

	viper.BindPFlag("path", discoverCmd.Flags().Lookup("path"))

	clientCmd.AddCommand(discoverCmd)
}

func RunDiscoverCmd(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("path")
	if path == "" {
		showerr(fmt.Errorf("no path specified"))
		return
	}
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(err)
		return
	}
	if path == "" {
		showerr(fmt.Errorf("missing path"))
		return
	}
	if err := c.DiscoverWithPath(path); err != nil {
		showerr(err)
	}
}
