package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	// root sfs cli command
	rootCmd = &cobra.Command{
		Use:   "sfs",
		Short: "root sfs command",
		Long: `
		Simple File Sync: synchronize your project files across your devices.

		Examples:
			sfs client --new
			sfs server --start
		`,
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
