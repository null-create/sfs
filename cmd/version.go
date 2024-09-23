package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "display installed sfs version",
		Run:   showVersion,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

func showVersion(cmd *cobra.Command, args []string) {
	b, err := os.ReadFile("VERSION")
	if err != nil {
		showerr(err)
		return
	}
	fmt.Print(string(b))
}
