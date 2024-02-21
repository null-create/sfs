package cmd

import (
	"fmt"
	"os"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ClientFlagPole struct {
	new     bool // create a new client
	start   bool // start a client
	local   bool // list all local files managed by SFS
	remote  bool // list all remote files managed by SFS
	refresh bool // refresh local drive
	info    bool // get information about the client
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
	clientCmd.PersistentFlags().BoolVar(&flags.info, "info", false, "Get info about the local SFS client")

	viper.BindPFlag("start", clientCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("stop", clientCmd.PersistentFlags().Lookup("stop"))
	viper.BindPFlag("new", clientCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("local", clientCmd.PersistentFlags().Lookup("local"))
	viper.BindPFlag("remote", clientCmd.PersistentFlags().Lookup("remote"))
	viper.BindPFlag("refresh", clientCmd.PersistentFlags().Lookup("refresh"))
	viper.BindPFlag("info", clientCmd.PersistentFlags().Lookup("info"))

	rootCmd.AddCommand(clientCmd)
}

func getflags(cmd *cobra.Command) ClientFlagPole {
	new, _ := cmd.Flags().GetBool("new")
	start, _ := cmd.Flags().GetBool("start")
	local, _ := cmd.Flags().GetBool("local")
	remote, _ := cmd.Flags().GetBool("remote")
	refresh, _ := cmd.Flags().GetBool("refresh")
	info, _ := cmd.Flags().GetBool("info")

	return ClientFlagPole{
		new:     new,
		start:   start,
		local:   local,
		remote:  remote,
		refresh: refresh,
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
	case f.refresh:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		c.RefreshDrive()
	case f.info:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(err)
		}
		fmt.Print(c.GetUserInfo())
	}
	return nil
}
