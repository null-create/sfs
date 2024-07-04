package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"
	"github.com/sfs/pkg/env"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Execute SFS Client Commands",
		RunE:  runClientCmd,
	}
	e = env.NewE()
)

func init() {
	flags := FlagPole{}
	clientCmd.PersistentFlags().BoolVar(&flags.new, "new", false, "Initialize a new client service instance")
	clientCmd.PersistentFlags().BoolVar(&flags.start, "start", false, "Start client services")
	clientCmd.PersistentFlags().BoolVar(&flags.info, "info", false, "Get info about the local SFS client")
	clientCmd.PersistentFlags().BoolVar(&flags.auto_sync, "auto-sync", true, "Enable or disable client auto-sync")

	viper.BindPFlag("start", clientCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("new", clientCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("info", clientCmd.PersistentFlags().Lookup("info"))
	viper.BindPFlag("auto-sync", clientCmd.PersistentFlags().Lookup("auto-sync"))

	rootCmd.AddCommand(clientCmd)
}

func getClientFlags(cmd *cobra.Command) FlagPole {
	new, _ := cmd.Flags().GetBool("new")
	start, _ := cmd.Flags().GetBool("start")
	info, _ := cmd.Flags().GetBool("info")
	autosync, _ := cmd.Flags().GetBool("auto-sync")

	return FlagPole{
		new:       new,
		start:     start,
		info:      info,
		auto_sync: autosync,
	}
}

func isNewService() (bool, error) {
	val, err := e.Get("CLIENT_NEW_SERVICE")
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

func runClientCmd(cmd *cobra.Command, args []string) error {
	f := getClientFlags(cmd)
	switch {
	case f.new:
		isNew, err := isNewService()
		if err != nil {
			return err
		}
		if !isNew {
			return fmt.Errorf("not a new instance")
		}
		_, err = client.Init(configs.NewService)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
		}
	case f.start:
		c, err := client.LoadClient(true)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
			return nil
		}
		err = c.Start()
		if err != nil {
			showerr(fmt.Errorf("failed to start client: %v", err))
		}
	case f.info:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
		}
		fmt.Print(c.GetUserInfo())
	case f.auto_sync || !f.auto_sync:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
		}
		c.SetAutoSync(f.auto_sync)
	}
	return nil
}
