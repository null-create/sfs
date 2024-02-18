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
	path string

	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push a file or directory to the SFS server",
		Run:   PushCmd,
	}
)

func init() {
	pushCmd.PersistentFlags().StringVar(&path, "path", "", "path to the file to push to the server")

	viper.BindPFlag("path", pushCmd.PersistentFlags().Lookup("push"))

	clientCmd.AddCommand(pushCmd)
}

func PushCmd(cmd *cobra.Command, args []string) {
	c, err := client.LoadClient(true)
	if err != nil {
		showerr(err)
	}
	filePath, _ := cmd.Flags().GetString("path")
	if filePath == "" {
		showerr(fmt.Errorf("please provide a path"))
		return
	}
	file, err := c.GetFileByPath(filePath)
	if err != nil {
		showerr(err)
		return
	}
	if err := c.PushFile(file); err != nil {
		showerr(err)
		return
	}
}
