package client

import (
	"math/rand"
	"os/exec"
	"runtime"
)

// Generate a pseudo-random integer in the range [0, n)
func RandInt(limit int) int {
	return rand.Intn(limit)
}

// open the web client home page in a new browser window
func Openbrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
