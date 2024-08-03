package cmd

import (
	"fmt"
	"time"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

/*
File for sending remote commands to the server.

"Remote" meaning the CLI can directly interface with the server to
execute commands and various upkeep.

sfs remote --is-up  // see if the server is up
sfs remote --stats  // see how long the server has been running plus some other stats?
*/

var (
	remoteCmd = &cobra.Command{
		Use:   "remote",
		Short: "Send remote commands to the server",
		Run:   runRemoteCmd,
	}
)

func init() {
	flags := FlagPole{}
	remoteCmd.Flags().BoolVar(&flags.isUp, "is-up", false, "Check whether the server is up")

	viper.BindPFlag("is-up", remoteCmd.PersistentFlags().Lookup("is-up"))
	viper.BindPFlag("stats", remoteCmd.PersistentFlags().Lookup("stats"))

	rootCmd.AddCommand(remoteCmd)
}

func getRemoteFlags(cmd *cobra.Command) FlagPole {
	isUp, _ := cmd.Flags().GetBool("is-up")
	return FlagPole{
		isUp: isUp,
	}
}

func secondsToTimeStr(seconds float64) string {
	duration := time.Duration(int64(seconds)) * time.Second
	timeValue := time.Time{}.Add(duration)
	return timeValue.Format("15:04:05")
}

func runRemoteCmd(cmd *cobra.Command, args []string) {
	f := getRemoteFlags(cmd)

	c, err := client.LoadClient(false)
	if err != nil {
		showerr(fmt.Errorf("failed to initialize service: %v", err))
	}

	if f.isUp {
		runtime, err := c.GetServerRuntime()
		if err != nil {
			showerr(err)
			return
		}
		fmt.Printf("server runtime: " + secondsToTimeStr(runtime))
	}
}
