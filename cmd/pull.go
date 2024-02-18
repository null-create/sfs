package cmd

import (
	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
Pull files or directories from the SFS server
*/

var (
	name string

	pullCmd = &cobra.Command{
		Use:   "pull",
		Short: "Pull files or directories from the SFS server",
		Run:   RunPullCmd,
	}
)

func init() {
	pullCmd.Flags().StringVar(&name, "path", "", "name of the item to pull")

	viper.BindPFlag("pull", pullCmd.PersistentFlags().Lookup("name"))

	clientCmd.AddCommand(pullCmd)
}

func RunPullCmd(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")

	c, err := client.LoadClient(false)
	if err != nil {
		showerr(err)
		return
	}

	file, err := c.GetFileByName(name)
	if err != nil {
		showerr(err)
		return
	}

	if err := c.PullFile(file); err != nil {
		showerr(err)
		return
	}
}
