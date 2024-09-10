package client

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// Generate a pseudo-random integer in the range [0, n)
// Used in tests.
func RandInt(limit int) int {
	return rand.Intn(limit)
}

// check if a file exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
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

// ShowFileInExplorer opens the file explorer window and highlights the specified file
func ShowFileInExplorer(filePath string) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("could not get absolute path: %v", err)
	}

	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", "/select,", absPath).Run() // Use explorer.exe to open the file location and highlight the file
	case "darwin":
		return exec.Command("open", "-R", absPath).Run() // Use the 'open' command on macOS to reveal the file in Finder
	case "linux":
		dir := filepath.Dir(absPath)
		return exec.Command("xdg-open", dir).Run() // Use 'xdg-open' to open the file's parent directory (does not highlight the file)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
