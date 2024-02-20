package cmd

import (
	"fmt"
	"os"

	"github.com/sfs/pkg/client"
	svc "github.com/sfs/pkg/service"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ClientFlagPole struct {
	new     bool   // create a new client
	start   bool   // start a client
	local   bool   // list all local files managed by SFS
	remote  bool   // list all remote files managed by SFS
	refresh bool   // refresh local drive
	add     string // add a file or directory to the local sfs service
	remove  string // remove a file or directory from the local sfs service
	info    bool   // get information about the client
}

var (
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Execute SFS Client Commands",
		RunE:  ClientCmd,
	}
)

func init() {
	flags := ClientFlagPole{}
	clientCmd.PersistentFlags().BoolVar(&flags.new, "new", false, "Initialize a new client service instance")
	clientCmd.PersistentFlags().BoolVar(&flags.start, "start", false, "Start client services")
	clientCmd.PersistentFlags().BoolVar(&flags.local, "local", false, "List local files managed by SFS service")
	clientCmd.PersistentFlags().BoolVar(&flags.remote, "remote", false, "List remote files managed by SFSService")
	clientCmd.PersistentFlags().BoolVar(&flags.refresh, "refresh", false, "Refresh drive. will search and add newly discovered files and directories")
	clientCmd.PersistentFlags().StringVar(&flags.add, "add", "", "Add a file to the local SFS filesystem. Pass file path to file to be added.")
	clientCmd.PersistentFlags().StringVar(&flags.remove, "remove", "", "Remove a file from the local SFS filesystem. Pass the file path of the file to be removed.")
	clientCmd.PersistentFlags().BoolVar(&flags.info, "info", false, "Get info about the local SFS client")

	viper.BindPFlag("start", clientCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("stop", clientCmd.PersistentFlags().Lookup("stop"))
	viper.BindPFlag("new", clientCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("local", clientCmd.PersistentFlags().Lookup("local"))
	viper.BindPFlag("remote", clientCmd.PersistentFlags().Lookup("remote"))
	viper.BindPFlag("push", clientCmd.PersistentFlags().Lookup("push"))
	viper.BindPFlag("pull", clientCmd.PersistentFlags().Lookup("pull"))
	viper.BindPFlag("add", clientCmd.PersistentFlags().Lookup("add"))
	viper.BindPFlag("remove", clientCmd.PersistentFlags().Lookup("remove"))
	viper.BindPFlag("info", clientCmd.PersistentFlags().Lookup("info"))

	rootCmd.AddCommand(clientCmd)
}

func getflags(cmd *cobra.Command) ClientFlagPole {
	new, _ := cmd.Flags().GetBool("new")
	start, _ := cmd.Flags().GetBool("start")
	local, _ := cmd.Flags().GetBool("local")
	remote, _ := cmd.Flags().GetBool("remote")
	refresh, _ := cmd.Flags().GetBool("refresh")
	add, _ := cmd.Flags().GetString("add")
	remove, _ := cmd.Flags().GetString("remove")
	info, _ := cmd.Flags().GetBool("info")

	return ClientFlagPole{
		new:     new,
		start:   start,
		local:   local,
		remote:  remote,
		refresh: refresh,
		add:     add,
		remove:  remove,
		info:    info,
	}
}

func ClientCmd(cmd *cobra.Command, args []string) error {
	f := getflags(cmd)
	switch {
	case f.new:
		_, err := client.Init(configs.NewService)
		if err != nil {
			showerr(err)
		}
	case f.start:
		c, err := client.LoadClient(true)
		if err != nil {
			showerr(err)
		}
		var shutdown = make(chan os.Signal)
		err = c.Start(shutdown)
		if err != nil {
			showerr(err)
		}
	case f.local:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		if err := c.ListLocalFilesDB(); err != nil {
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
	case f.add != "": // NOTE add will have the path of the file to be added
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		// determine item type, then add
		item, err := os.Stat(f.add)
		if err != nil {
			return err
		}
		// NOTE: both newFile.DirID == "" and newDir.ID == "" here.
		if item.IsDir() {
			newDir := svc.NewDirectory(item.Name(), c.UserID, c.DriveID, f.add)
			if err := c.AddDirWithID(newDir.ID, newDir); err != nil {
				showerr(err)
			}
		} else {
			newFile := svc.NewFile(item.Name(), c.DriveID, c.UserID, f.add)
			// TODO: need to have a directory for this file to go to, even
			// if it's root. newFile.DirID == "" in this current implementation.
			//
			// TODO: decide whether:
			// - a file should only be monitored if it's in the designated root
			//   a file would be registed by passing its current path in the
			//   users machine, and would be copied to the root directory if we
			//   choose to add it to the service. any future modifications would
			//   have to to be done in the version within the sfs root directory.
			//
			// - a file can be added by the passing its current path in the users
			//   machine and be monitored *without* having to move it to a
			//   sfs root directory.
			if err := c.AddFileWithID(newFile.DirID, newFile); err != nil {
				showerr(err)
			}
		}
	case f.remove != "":
		return nil
	case f.refresh:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		if err := c.RefreshDrive(); err != nil {
			showerr(err)
		}
	case f.info:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		fmt.Print(c.GetUserInfo())
	}
	return nil
}
