package cmd

import (
	"fmt"

	"github.com/sfs/pkg/logger"
)

var cmdLogger = logger.NewLogger("CLI")

func showerr(err error) {
	cmdLogger.Error(err.Error())
	fmt.Print("\n" + err.Error() + "\n")
}
