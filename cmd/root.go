package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/sfs/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	c *client.Client // active client service instance

	cfgFile string

	// root sfs cli command
	rootCmd = &cobra.Command{
		Use:     "sfs",
		Short:   "Root SFS Command. Use to call client and server commands.",
		Version: "0.1", // TODO: better semantic versioning
	}
)

func init() {
	initConfig()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
