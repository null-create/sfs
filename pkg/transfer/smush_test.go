package transfer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sfs/pkg/env"
)

func TestCompress(t *testing.T) {
	env.SetEnv(false)

	tmpDir, err := MakeTmpTxtFiles(t, GetTestingDir())
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	if err := Zip(tmpDir.Path, filepath.Join(tmpDir.Path, "test.zip")); err != nil {
		Fail(t, GetTestingDir(), fmt.Errorf("failed to compress test dir: %v", err))
	}

	entries, err := os.ReadDir(tmpDir.Path)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if !strings.Contains(entries[0].Name(), ".zip") {
		Fail(t, GetTestingDir(), fmt.Errorf("archive folder not found"))
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}

func TestUnpack(t *testing.T) {
	env.SetEnv(false)

	tmpDir, err := MakeTmpTxtFiles(t, GetTestingDir())
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}

	testArchive := filepath.Join(tmpDir.Path, "test.zip")
	if err := Zip("testing", testArchive); err != nil {
		Fail(t, GetTestingDir(), fmt.Errorf("failed to compress test dir: %v", err))
	}

	entries, err := os.ReadDir(tmpDir.Path)
	if err != nil {
		Fail(t, GetTestingDir(), err)
	}
	if len(entries) < 11 {
		Fail(t, GetTestingDir(), fmt.Errorf("zip file not created"))
	}

	if err := Unzip(testArchive, filepath.Join("testing", "tmp-result")); err != nil {
		Fail(t, GetTestingDir(), err)
	}

	if err := Clean(t, GetTestingDir()); err != nil {
		t.Fatal(err)
	}
}
