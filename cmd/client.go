package cmd

import (
	"log"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ccfg = client.ClientConfig() // client service configurations
	sig  chan bool               // client shut down signal

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
				newClient, err := client.Init(ccfg.NewService)
				if err != nil {
					return err
				}
				c = newClient
			case startFlag:
				client, err := client.LoadClient()
				if err != nil {
					return err
				}
				c = client
				go func() {
					off, err := c.Start()
					if err != nil {
						log.Printf("failed to start client: %v", err)
						return
					}
					sig = off
				}()
			case stopFlag:
				sig <- true
				return c.ShutDown()
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
