package cmd

import "fmt"

func showerr(err error) {
	fmt.Print("\n" + err.Error() + "\n")
}
