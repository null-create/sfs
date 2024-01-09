package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	local  string
	remote string
	dflt   = ""

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "list all local and remote files and directories managed by SFS",
		RunE: func(cmd *cobra.Command, args []string) error {
			local, _ := cmd.Flags().GetString("local")
			remote, _ := cmd.Flags().GetString("remote")
			switch {
			case local != dflt:
				c.ListLocalFiles()
			case remote != dflt:
				if err := c.ListRemoteFiles(); err != nil {
					return err
				}
			}
			return nil
		},
	}
)

func init() {
	listCmd.PersistentFlags().StringVarP(&local, "local", "l", "local", "list all local files managed by SFS")
	listCmd.PersistentFlags().StringVarP(&remote, "remote", "r", "remote", "list all remote files managed by SFS")

	viper.BindPFlag("local", listCmd.PersistentFlags().Lookup("local"))
	viper.BindPFlag("remote", listCmd.PersistentFlags().Lookup("remote"))

	clientCmd.AddCommand(listCmd)
}
