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
	stop    bool   // stop a client
	local   bool   // list all local files managed by SFS
	remote  bool   // list all remote files managed by SFS
	refresh bool   // refresh local drive
	sync    bool   // sync with the server
	push    string // push a file to the server
	pull    string // pull a file from the server
	add     string // add a file or directory to the local sfs service
	remove  string // remove a file or directory from the local sfs service
	info    bool   // get information about the client
}

var (
	shutdown chan os.Signal // shutdown signal

	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Execute SFS Client Commands",
		RunE:  RunClientCmd,
	}
)

func init() {
	flags := ClientFlagPole{}
	clientCmd.PersistentFlags().BoolVar(&flags.new, "new", false, "Initialize a new client service instance")
	clientCmd.PersistentFlags().BoolVar(&flags.start, "start", false, "Start client services")
	clientCmd.PersistentFlags().BoolVar(&flags.stop, "stop", false, "Stop client services")
	clientCmd.PersistentFlags().BoolVar(&flags.local, "local", false, "List local files managed by SFS service")
	clientCmd.PersistentFlags().BoolVar(&flags.remote, "remote", false, "List remote files managed by SFSService")
	clientCmd.PersistentFlags().BoolVar(&flags.refresh, "refresh", false, "Refresh drive. will search and add newly discovered files and directories")
	clientCmd.PersistentFlags().BoolVar(&flags.sync, "sync", false, "Sync with the remote server")
	clientCmd.PersistentFlags().StringVar(&flags.push, "push", "", "Push a file to the remote server. Add path to flag.")
	clientCmd.PersistentFlags().StringVar(&flags.pull, "pull", "", "Pull a file from the remote server. Add filename to flag.")
	clientCmd.PersistentFlags().StringVar(&flags.add, "add", "", "Add a file to the local SFS filesystem. Pass file path to file to be added.")
	clientCmd.PersistentFlags().StringVar(&flags.remove, "remove", "", "Remove a file from the local SFS filesystem. Pass the file path of the file to be removed.")
	clientCmd.PersistentFlags().BoolVar(&flags.info, "info", false, "Get info about the local SFS client")

	viper.BindPFlag("start", clientCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("stop", clientCmd.PersistentFlags().Lookup("stop"))
	viper.BindPFlag("new", clientCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("local", clientCmd.PersistentFlags().Lookup("local"))
	viper.BindPFlag("remote", clientCmd.PersistentFlags().Lookup("remote"))
	viper.BindPFlag("sync", clientCmd.PersistentFlags().Lookup("sync"))
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
	stop, _ := cmd.Flags().GetBool("stop")
	local, _ := cmd.Flags().GetBool("local")
	remote, _ := cmd.Flags().GetBool("remote")
	sync, _ := cmd.Flags().GetBool("sync")
	push, _ := cmd.Flags().GetString("push")
	pull, _ := cmd.Flags().GetString("pull")
	refresh, _ := cmd.Flags().GetBool("refresh")
	add, _ := cmd.Flags().GetString("add")
	remove, _ := cmd.Flags().GetString("remove")
	info, _ := cmd.Flags().GetBool("info")

	return ClientFlagPole{
		new:     new,
		start:   start,
		stop:    stop,
		local:   local,
		remote:  remote,
		sync:    sync,
		push:    push,
		pull:    pull,
		refresh: refresh,
		add:     add,
		remove:  remove,
		info:    info,
	}
}

func RunClientCmd(cmd *cobra.Command, args []string) error {
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
		off, err := c.Start()
		if err != nil {
			showerr(err)
		}
		shutdown = off
	case f.stop:
		shutdown <- os.Kill
	case f.local:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		if err := c.ListLocalFilesDB(); err != nil {
			showerr(err)
		}
	case f.remote:
		c, err := client.LoadClient(true)
		if err != nil {
			showerr(err)
		}
		if err := c.ListRemoteFiles(); err != nil {
			showerr(err)
		}
	case f.sync:
		// TODO:
	case f.push != "":
		c, err := client.LoadClient(true)
		if err != nil {
			showerr(err)
		}
		// push file
		var path = f.push
		file, err := c.GetFileByPath(path)
		if err != nil {
			showerr(err)
		}
		if err := c.PushFile(file); err != nil {
			showerr(err)
		}
	case f.pull != "":
		c, err := client.LoadClient(true)
		if err != nil {
			showerr(err)
		}
		// pull files
		var path = f.pull
		file, err := c.GetFileByPath(path)
		if err != nil {
			showerr(err)
		}
		if err := c.PullFile(file); err != nil {
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
			if err := c.AddDir(newDir.ID, newDir); err != nil {
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
			if err := c.AddFile(newFile.DirID, newFile); err != nil {
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
