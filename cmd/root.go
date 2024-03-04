package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgfile string

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
	if cfgfile != "" {
		viper.SetConfigFile(cfgfile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		cmdLogger.Info(fmt.Sprintf("Using config file: %v", viper.ConfigFileUsed()))
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		cmdLogger.Error(fmt.Sprintf("failed to execute root command: %v", err))
		log.Fatal(err)
	}
}
