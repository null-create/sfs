package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Push a file or directory to the SFS server
*/

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push a file or directory to the SFS server",
		Run:   PushCmd,
	}
)

func init() {
	flags := FlagPole{}
	pushCmd.PersistentFlags().StringVar(&flags.path, "path", "", "path to the file to push to the server")
	pushCmd.PersistentFlags().BoolVarP(&flags.isDir, "is-dir", "d", false, "flag for whether we're sending a directory.")
	pushCmd.PersistentFlags().BoolVar(&flags.newFile, "new-file", false, "flag for whether this is a new file to the service")
	pushCmd.PersistentFlags().BoolVar(&flags.newDir, "new-dir", false, "flag for whether this is a new directory to the service")

	viper.BindPFlag("path", pushCmd.PersistentFlags().Lookup("path"))
	viper.BindPFlag("new-file", pushCmd.PersistentFlags().Lookup("new-file"))
	viper.BindPFlag("new-dir", pushCmd.PersistentFlags().Lookup("new-dir"))
	viper.BindPFlag("is-dir", pushCmd.PersistentFlags().Lookup("is-dir"))

	clientCmd.AddCommand(pushCmd)
}

func PushCmd(cmd *cobra.Command, args []string) {
	filePath, _ := cmd.Flags().GetString("path")
	if filePath == "" {
		showerr(fmt.Errorf("no file path specified"))
		return
	}
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(fmt.Errorf("failed to initialize service: %v", err))
	}
	file, err := c.GetFileByPath(filePath)
	if err != nil {
		showerr(fmt.Errorf("failed to get file: %v", err))
		return
	}
	newFile, _ := cmd.Flags().GetBool("new-file")
	if newFile {
		if err := c.PushNewFile(file); err != nil {
			showerr(fmt.Errorf("failed to push new file: %v", err))
		}
	} else {
		if err := c.PushFile(file); err != nil {
			showerr(fmt.Errorf("failed to push file: %v", err))
		}
	}
}
