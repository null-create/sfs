package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Pull files or directories from the SFS server
*/

var (
	pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "Pull files or directories from the SFS server",
		Run:   runPullCmd,
	}
)

func init() {
	flags := FlagPole{}
	pullCmd.Flags().StringVar(&flags.name, "name", "", "name of the file or directory to pull from the server")

	viper.BindPFlag("pull", pullCmd.PersistentFlags().Lookup("name"))

	drvCmd.AddCommand(pullCmd)
}

func runPullCmd(cmd *cobra.Command, args []string) {
	// see if local backup mode is enabled first
	// if so, thent he client won't have files stored on the sfs server
	if localBackupIsEnabled() {
		fmt.Print("local backup mode is enabled. remote files are not available.")
		return
	}
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		showerr(fmt.Errorf("no name specified"))
		return
	}

	c, err := client.LoadClient(false)
	if err != nil {
		showerr(fmt.Errorf("failed to initialize service: %v", err))
		return
	}

	file, err := c.GetFileByName(name)
	if err != nil {
		showerr(fmt.Errorf("failed to get file: %v", err))
		return
	}
	if file == nil {
		showerr(fmt.Errorf("file %s not found", name))
		return
	}

	if err := c.PullFile(file); err != nil {
		showerr(fmt.Errorf("failed to pull file: %v", err))
		return
	}
}
