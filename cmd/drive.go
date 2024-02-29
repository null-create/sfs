package cmd

import (
	"github.com/sfs/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Command for registering a new drive and user with the server
as well as other drive management commands

sfs drive --new
sfs drive --register
sfs drive --refresh
sfs drive --list-files
sfs drive --list-dirs

// add or remove files

sfs drive add --path
sfs drive remove --path
*/

var (
	drvCmd = &cobra.Command{
		Use:   "drive",
		Short: "Command for managing the client side drive service",
		Run:   RunDrvCmd,
	}
)

func init() {
	flags := FlagPole{}
	drvCmd.Flags().BoolVar(&flags.register, "register", false, "register a new drive with the sfs server")
	drvCmd.Flags().BoolVar(&flags.list_files, "list-files", false, "list all local files managed by the sfs client service")
	drvCmd.Flags().BoolVar(&flags.list_dirs, "list-dirs", false, "list all local directories managed by the sfs client service")
	drvCmd.Flags().BoolVar(&flags.remote, "remote", false, "list all files stored on the sfs server")

	viper.BindPFlag("register", drvCmd.PersistentFlags().Lookup("register"))
	viper.BindPFlag("list-files", drvCmd.PersistentFlags().Lookup("list-files"))
	viper.BindPFlag("list-dirs", drvCmd.PersistentFlags().Lookup("list-dirs"))
	viper.BindPFlag("remote", drvCmd.Flags().Lookup("remote"))

	rootCmd.AddCommand(drvCmd)
}

func getDrvflags(cmd *cobra.Command) FlagPole {
	register, _ := cmd.Flags().GetBool("register")
	list_files, _ := cmd.Flags().GetBool("list-files")
	list_dirs, _ := cmd.Flags().GetBool("list-dirs")
	remote, _ := cmd.Flags().GetBool("remote")

	return FlagPole{
		register:   register,
		list_files: list_files,
		list_dirs:  list_dirs,
		remote:     remote,
	}
}

func RunDrvCmd(cmd *cobra.Command, args []string) {
	f := getDrvflags(cmd)
	switch {
	case f.register:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		if err := c.RegisterDrive(); err != nil {
			showerr(err)
		}
	case f.list_files:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		if err := c.ListLocalFilesDB(); err != nil {
			showerr(err)
		}
	case f.list_dirs:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		if err := c.ListLocalDirsDB(); err != nil {
			showerr(err)
		}
	case f.remote:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		if err := c.ListRemoteFiles(); err != nil {
			showerr(err)
		}
	case f.refresh:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		c.RefreshDrive()
	}
}
