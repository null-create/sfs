package cmd

import (
	"fmt"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	configs   = client.GetClientConfigs()
	clientCmd = &cobra.Command{
		Use:   "client",
		Short: "SFS client commands",
		Run:   runClientCmd,
	}
)

// TODO: add support for launching the web UI (when its ready)

func init() {
	flags := FlagPole{}
	clientCmd.PersistentFlags().BoolVar(&flags.new, "new", false, "Initialize a new client service instance")
	clientCmd.PersistentFlags().BoolVar(&flags.info, "info", false, "Get info about the local SFS client")
	clientCmd.PersistentFlags().BoolVar(&flags.register, "register", false, "Register the client with the server")
	clientCmd.PersistentFlags().BoolVar(&flags.browser, "browser", false, "Launch the browser interface")

	viper.BindPFlag("new", clientCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("info", clientCmd.PersistentFlags().Lookup("info"))
	viper.BindPFlag("register", clientCmd.PersistentFlags().Lookup("register"))
	viper.BindPFlag("browser", clientCmd.PersistentFlags().Lookup("browser"))

	rootCmd.AddCommand(clientCmd)
}

func getClientFlags(cmd *cobra.Command) FlagPole {
	new, _ := cmd.Flags().GetBool("new")
	info, _ := cmd.Flags().GetBool("info")
	register, _ := cmd.Flags().GetBool("register")
	browser, _ := cmd.Flags().GetBool("browser")

	return FlagPole{
		new:      new,
		info:     info,
		register: register,
		browser:  browser,
	}
}

func isNewService() (bool, error) {
	val, err := envCfgs.Get("CLIENT_NEW_SERVICE")
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

func runClientCmd(cmd *cobra.Command, args []string) {
	f := getClientFlags(cmd)
	switch {
	case f.new:
		isNew, err := isNewService()
		if err != nil {
			showerr(err)
			return
		}
		if !isNew {
			showerr(fmt.Errorf("not a new instance"))
			return
		}
		_, err = client.Init(configs.NewService)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize new service: %v", err))
		}
	case f.register:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
		}
		if err := c.RegisterClient(); err != nil {
			showerr(err)
		}

	case f.info:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
			return
		}
		fmt.Print(c.GetUserInfo())
	}
}
