package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sfs/pkg/client"

	"github.com/spf13/cobra"
)

// command to empty the SFS recycle bin

var (
	cleanCmd = &cobra.Command{
		Use:   "clean",
		Short: "Command to empty the SFS recycle bin.",
		Run:   cleanRecycleBin,
	}
)

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func proceed(totalEntries int) bool {
	var answer string
	fmt.Printf("WARNING: this process will delete %d files in the SFS recycle bin. Do you want to proceed? (y/n)", totalEntries)
	fmt.Scan(&answer)
	return strings.ToLower(answer) == "y"
}

func cleanRecycleBin(cmd *cobra.Command, args []string) {
	c, err := client.LoadClient(false)
	if err != nil {
		showerr(fmt.Errorf("failed to initialize service: %v", err))
		return
	}
	// get input confirmation before proceeding
	entries, err := os.ReadDir(c.RecycleBin)
	if err != nil {
		showerr(err)
		return
	}
	if len(entries) == 0 {
		return
	}
	if !proceed(len(entries)) {
		return
	}
	// clean the SFS recycle bin
	if err := c.EmptyRecycleBin(); err != nil {
		showerr(err)
	}
}
