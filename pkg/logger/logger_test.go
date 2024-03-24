package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/env"
)

func TestLogger(t *testing.T) {
	env.SetEnv(false)

	logger := NewLogger("TEST_LOGGER", "None")

	logger.Info("test info message")
	logger.Warn("another test message")
	logger.Debug("yet another test message")
	logger.Error("yet another test message")

	entries, err := os.ReadDir(filepath.Dir(logger.logfile))
	if err != nil {
		Fatal(t, err)
	}
	if len(entries) != 1 {
		Fail(t, GetTestingDir(), fmt.Errorf("no log file found"))
	}
}
