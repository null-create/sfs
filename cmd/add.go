package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Command for adding a new file or directory to the local SFS client service
*/

var (
	addCmd = &cobra.Command{
		Use:   "add",
		Short: "Add a new file or directory to the local SFS client service",
		Run:   runAddCmd,
	}
)

func init() {
	flags := FlagPole{}
	addCmd.Flags().StringVarP(&flags.path, "path", "p", "", "Path to the new file or directory")

	viper.BindPFlag("path", addCmd.Flags().Lookup("path"))

	drvCmd.AddCommand(addCmd)
}

func runAddCmd(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("path")
	if path == "" {
		showerr(fmt.Errorf("no path specified"))
		return
	}
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(fmt.Errorf(fmt.Sprintf("failed to load client: %v", err)))
		return
	}
	if err := c.AddItem(path); err != nil {
		showerr(fmt.Errorf(fmt.Sprintf("failed to add item: %v", err)))
	}
}
