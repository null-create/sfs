package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Command for managing the clients user

NOTE: sfs update handles client configuration updates

sfs user --new            // get a users info from the command line
sfs user --info           // show the client info
sfs user --remove         // remove all clients info (files, directories, etc...)
sfs user --is-admin --key // set this user as an admin.
*/

var (
	userCmd = &cobra.Command{
		Use:   "user",
		Short: "Commands for managing the clients info",
		Run:   runUserCmd,
	}
)

func init() {
	flags := FlagPole{}
	userCmd.PersistentFlags().BoolVar(&flags.new, "new", false, "add a new user")
	userCmd.PersistentFlags().BoolVar(&flags.info, "info", false, "whether to retrieve info about a user")
	userCmd.PersistentFlags().BoolVar(&flags.remove, "remove", false, "whether we should remove this user")

	viper.BindPFlag("new", userCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("info", userCmd.PersistentFlags().Lookup("info"))
	viper.BindPFlag("remove", userCmd.PersistentFlags().Lookup("remove"))

	rootCmd.AddCommand(userCmd)
}

func getUserFlags(cmd *cobra.Command) FlagPole {
	new, _ := cmd.Flags().GetBool("new")
	info, _ := cmd.Flags().GetBool("info")
	remove, _ := cmd.Flags().GetBool("remove")

	return FlagPole{
		new:    new,
		info:   info,
		remove: remove,
	}
}

func runUserCmd(cmd *cobra.Command, args []string) {
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(fmt.Errorf("failed to initialize service: %v", err))
	}

	f := getUserFlags(cmd)

	switch {
	case f.new:
		if err := c.AddNewUser(); err != nil {
			showerr(err)
		}
	case f.info:
		fmt.Print(c.GetUserInfo())
	case f.remove:
		if err := c.RemoveUser(); err != nil {
			showerr(err)
		}
	}
}
