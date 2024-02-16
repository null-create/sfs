package cmd

import (
	"os"

	"github.com/sfs/pkg/client"
	svc "github.com/sfs/pkg/service"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	c        *client.Client // active client service instance
	shutdown chan os.Signal // shutdown signal
	configs  = client.ClientConfig()

	// flags
	newClient    bool   // create a new client
	startClient  bool   // start a client
	stopClient   bool   // stop a client
	listLoc      bool   // list all local files managed by SFS
	listRemote   bool   // list all remote files managed by SFS
	refreshDrive bool   // refresh local drive
	sync         bool   // sync with the server
	push         string // push a file to the server
	pull         string // pull a file from the server
	add          string // add a file or directory to the local sfs service
	remove       string // remove a file or directory from the local sfs service

	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Execute SFS Client Commands",
		RunE:  RunCmd,
	}
)

func init() {
	clientCmd.PersistentFlags().BoolVar(&newClient, "new", false, "Initialize a new client service instance")
	clientCmd.PersistentFlags().BoolVar(&startClient, "start", false, "Start client services")
	clientCmd.PersistentFlags().BoolVar(&stopClient, "stop", false, "Stop client services")
	clientCmd.PersistentFlags().BoolVar(&listLoc, "local", false, "List local files managed by SFS service")
	clientCmd.PersistentFlags().BoolVar(&listRemote, "remote", false, "List remote files managed by SFSService")
	clientCmd.PersistentFlags().BoolVar(&refreshDrive, "refresh", false, "Refresh drive. will search and add newly discovered files and directories")
	clientCmd.PersistentFlags().BoolVar(&sync, "sync", false, "Sync with the remote server")
	clientCmd.PersistentFlags().StringVar(&push, "push", "", "Push a file to the remote server. Add path to flag.")
	clientCmd.PersistentFlags().StringVar(&pull, "pull", "", "Pull a file from the remote server. Add filename to flag.")
	clientCmd.PersistentFlags().StringVar(&add, "add", "", "Add a file to the local SFS filesystem. Pass file path to file to be added.")
	clientCmd.PersistentFlags().StringVar(&remove, "remove", "", "Remove a file from the local SFS filesystem. Pass the file path of the file to be removed.")

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

	rootCmd.AddCommand(clientCmd)
}

func RunCmd(cmd *cobra.Command, args []string) error {
	new, _ := cmd.Flags().GetBool("new")
	start, _ := cmd.Flags().GetBool("start")
	stop, _ := cmd.Flags().GetBool("stop")
	local, _ := cmd.Flags().GetBool("local")
	remote, _ := cmd.Flags().GetBool("remote")
	sync, _ := cmd.Flags().GetBool("sync")
	push, _ := cmd.Flags().GetString("push")
	pull, _ := cmd.Flags().GetString("pull")
	add, _ := cmd.Flags().GetString("add")
	remove, _ := cmd.Flags().GetString("remove")

	switch {
	case new:
		_, err := client.Init(configs.NewService)
		if err != nil {
			return err
		}
	case start:
		c, err := client.LoadClient()
		if err != nil {
			return err
		}
		off, err := c.Start()
		if err != nil {
			return err
		}
		shutdown = off
	case stop:
		shutdown <- os.Kill
	case local:
		c, err := client.LoadClient()
		if err != nil {
			return err
		}
		c.ListLocalFiles()
	case remote:
		c, err := client.LoadClient()
		if err != nil {
			return err
		}
		if err := c.ListRemoteFiles(); err != nil {
			return err
		}
	case sync:
		// TODO:
	case push != "":
		c, err := client.LoadClient()
		if err != nil {
			return err
		}
		// push file
		var path = push
		file, err := c.GetFileByPath(path)
		if err != nil {
			return err
		}
		if err := c.PushFile(file); err != nil {
			return err
		}
	case pull != "":
		c, err := client.LoadClient()
		if err != nil {
			return err
		}
		// pull files
		var path = pull
		file, err := c.GetFileByPath(path)
		if err != nil {
			return err
		}
		if err := c.PullFile(file); err != nil {
			return err
		}
	case add != "":
		c, err := client.LoadClient()
		if err != nil {
			return err
		}
		// determine item type, then add
		item, err := os.Stat(add)
		if err != nil {
			return err
		}
		// NOTE: both newFile.DirID == "" and newDir.ID == "" here.
		if item.IsDir() {
			newDir := svc.NewDirectory(item.Name(), c.UserID, c.DriveID, add)
			if err := c.AddDir(newDir.ID, newDir); err != nil {
				return err
			}
		} else {
			newFile := svc.NewFile(item.Name(), c.DriveID, c.UserID, add)
			if err := c.AddFile(newFile.DirID, newFile); err != nil {
				return err
			}
		}
	case remove != "":
	}

	return nil
}
