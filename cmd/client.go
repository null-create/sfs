package cmd

import (
	"os"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clnt *client.Client // active client service instance

	shutdown chan os.Signal

	configs = client.ClientConfig()

	// for flags
	newClient   bool
	startClient bool
	stopClient  bool

	// main client command
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "SFS Client Commands",
		RunE: func(cmd *cobra.Command, args []string) error {
			newFlag, _ := cmd.Flags().GetBool("new")
			startFlag, _ := cmd.Flags().GetBool("start")
			stopFlag, _ := cmd.Flags().GetBool("stop")

			switch {
			case newFlag:
				_, err := client.Init(configs.NewService)
				if err != nil {
					return err
				}
			case startFlag:
				off, err := clnt.Start()
				if err != nil {
					return err
				}
				shutdown = off
			case stopFlag:
				shutdown <- os.Kill // stops blocking client process
			}
			return nil
		},
	}
)

func init() {
	clientCmd.PersistentFlags().BoolVar(&newClient, "new", false, "Initialize a new client service instance")
	clientCmd.PersistentFlags().BoolVar(&startClient, "start", false, "Start client services")
	clientCmd.PersistentFlags().BoolVar(&stopClient, "stop", false, "Stop client services")

	viper.BindPFlag("start", clientCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("stop", clientCmd.PersistentFlags().Lookup("stop"))
	viper.BindPFlag("new", clientCmd.PersistentFlags().Lookup("new"))

	rootCmd.AddCommand(clientCmd)
}
