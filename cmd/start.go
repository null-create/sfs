package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start sfs Client. Use --browser flag to run in browser mode.",
		Run:   runStartCmd,
	}
)

func init() {
	flags := FlagPole{}
	startCmd.PersistentFlags().BoolVarP(&flags.browser, "browser", "b", false, "Run in browser mode")

	viper.BindPFlag("browser", startCmd.PersistentFlags().Lookup("browser"))

	clientCmd.AddCommand(startCmd)
}

func getStartFlag(cmd *cobra.Command) FlagPole {
	browser, _ := cmd.Flags().GetBool("browser")
	return FlagPole{
		browser: browser,
	}
}

func runStartCmd(cmd *cobra.Command, args []string) {
	f := getStartFlag(cmd)

	c, err := client.LoadClient(true)
	if err != nil {
		showerr(fmt.Errorf("failed to initialize service: %v", err))
		return
	}

	if f.browser {
		c.StartWebClient() // run in the browser
	} else {
		c.Start() // run in the terminal
	}
}
