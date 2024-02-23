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
		Run:   RunPullCmd,
	}
)

func init() {
	flags := FlagPole{}
	pullCmd.Flags().StringVar(&flags.name, "name", "", "name of the item to pull")

	viper.BindPFlag("pull", pullCmd.PersistentFlags().Lookup("name"))

	clientCmd.AddCommand(pullCmd)
}

func RunPullCmd(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("name")
	if name == "" {
		showerr(fmt.Errorf("no name specified"))
		return
	}

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
	if file == nil {
		showerr(fmt.Errorf("file %s not found", name))
		return
	}

	if err := c.PullFile(file); err != nil {
		showerr(err)
		return
	}
}
