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
		Run:   runDiscoverCmd,
	}
)

func init() {
	flags := FlagPole{}
	discoverCmd.Flags().StringVarP(&flags.path, "path", "p", "", "Path to the directory to run discover on")

	viper.BindPFlag("path", discoverCmd.Flags().Lookup("path"))

	rootCmd.AddCommand(discoverCmd)
}

func runDiscoverCmd(cmd *cobra.Command, args []string) {
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
	if _, err := c.DiscoverWithPath(path); err != nil {
		showerr(err)
	}
}
