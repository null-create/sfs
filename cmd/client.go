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

func init() {
	flags := FlagPole{}
	clientCmd.PersistentFlags().BoolVar(&flags.new, "new", false, "Initialize a new client service instance")
	clientCmd.PersistentFlags().BoolVar(&flags.start, "start", false, "Start client services")
	clientCmd.PersistentFlags().BoolVar(&flags.info, "info", false, "Get info about the local SFS client")
	clientCmd.PersistentFlags().BoolVar(&flags.register, "register", false, "Register the client with the server")

	viper.BindPFlag("start", clientCmd.PersistentFlags().Lookup("start"))
	viper.BindPFlag("new", clientCmd.PersistentFlags().Lookup("new"))
	viper.BindPFlag("info", clientCmd.PersistentFlags().Lookup("info"))
	viper.BindPFlag("register", clientCmd.PersistentFlags().Lookup("register"))

	rootCmd.AddCommand(clientCmd)
}

func getClientFlags(cmd *cobra.Command) FlagPole {
	new, _ := cmd.Flags().GetBool("new")
	start, _ := cmd.Flags().GetBool("start")
	info, _ := cmd.Flags().GetBool("info")
	register, _ := cmd.Flags().GetBool("register")

	return FlagPole{
		new:      new,
		start:    start,
		info:     info,
		register: register,
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
	case f.start:
		c, err := client.LoadClient(true)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
			return
		}
		c.Start() // starts a blocking process to allow services to run
	case f.info:
		c, err := client.LoadClient(false)
		if err != nil {
			showerr(fmt.Errorf("failed to initialize service: %v", err))
			return
		}
		fmt.Print(c.GetUserInfo())
	}
}
