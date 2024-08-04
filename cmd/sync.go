package cmd

import (
	"github.com/sfs/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
File for initiating sync operations with the server
*/

var (
	syncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Invoke client synchronization operations",
		Long: `
Invoke client synchronization operations, either with local backups, or with a remote server.

Examples:
	// checks all monitored files and directories, and updates local dabases
	
	sfs drive sync --local=true 

	// calls server for latest metadata about monitored files and directories
	// then compares against the local filesystem. client will then attempt to
	// update the local filesystem and/or server with latest metadata and file objects. 

	sfs drive sync --remote=true 		
`,
		Run: runSyncCmd,
	}
)

func init() {
	flags := FlagPole{}

	syncCmd.Flags().BoolVar(&flags.local, "local", false, "sync local objects")
	syncCmd.Flags().BoolVar(&flags.remote, "remote", false, "sync remote objects")

	viper.BindPFlag("local", syncCmd.Flags().Lookup("local"))
	viper.BindPFlag("remote", syncCmd.Flags().Lookup("remote"))

	drvCmd.AddCommand(syncCmd)
}

func getFlags(cmd *cobra.Command) FlagPole {
	local, _ := cmd.Flags().GetBool("local")
	remote, _ := cmd.Flags().GetBool("remote")
	return FlagPole{
		local:  local,
		remote: remote,
	}
}

func runSyncCmd(cmd *cobra.Command, args []string) {
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(err)
	}
	f := getFlags(cmd)
	switch {
	case f.local:
		if err := c.LocalSync(); err != nil {
			showerr(err)
		}
	case f.remote:
		if err := c.ServerSync(); err != nil {
			showerr(err)
		}
	}
}
