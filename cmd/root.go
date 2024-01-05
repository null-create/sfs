package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sfs",
	Short: "root sfs command",
	// long description for help messages
	Long: `
	A bunch of text can go here, along with some demos about how
	to use the command.
	
	
	`,
	// filler/demo func for how to use this command
	Run: func(cmd *cobra.Command, args []string) {
		// whatever the command should call or do
	},
}

func init() {
	rootCmd.AddCommand(rootCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
