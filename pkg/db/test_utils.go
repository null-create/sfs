package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

const txtData string = "all work and no play makes jack a dull boy\n"

// handle test failures
//
// similar to Fatal(), except you can supply a
// testing/tmp directy path to clean
func Fail(t *testing.T, dir string, err error) {
	if err := Clean(t, dir); err != nil {
		log.Fatal(err)
	}
	t.Fatalf("[ERROR] %v", err)
}

// handles test failures,
// supplies [ERROR] prefix to supplied error messages
//
// it's just a call to Clean() followed by
// a call to t.Fatalf()
func Fatal(t *testing.T, err error) {
	if err2 := Clean(t, GetTestingDir()); err2 != nil {
		log.Printf("[ERROR] failed to clean testing directory during recovery: %v", err2)
	}
	t.Fatalf("[ERROR] %v", err)
}

func GetTestingDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("[WARNING] unable to get testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0666); err != nil {
			log.Fatalf("[ERROR] unable to create test directory: %v", err)
		}
	}
	return filepath.Join(curDir, "testing")
}

// make a temp .txt file of n size (in bytes).
//
// n is determined by textReps since that will be how
// many times testData is written to the text file
//
// returns a file pointer to the new temp file
func MakeTmpTxtFile(filePath string, textReps int) (*svc.File, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	var data string
	f := svc.NewFile(filepath.Base(filePath), "some-rand-id", "me", filePath)
	for i := 0; i < textReps; i++ {
		data += txtData
	}
	if err = f.Save([]byte(data)); err != nil {
		return nil, err
	}
	return f, nil
}

// make an empty test dir object. does not create a physical directory.
func MakeTestDir(path string) *svc.Directory {
	return svc.NewDirectory("bill", "bill buttlicker", "some-rand-id", filepath.Join(path, "bill"))
}

// make a tmp user for testing
func MakeTestUser(testDir string) *auth.User {
	return auth.NewUser("bill buttlicker", "billb", "bill@bill.com", filepath.Join(testDir, "bill"), false)
}

func MakeTestItems(t *testing.T, testDir string) (*svc.Drive, *svc.Directory, *auth.User) {
	tempDir := svc.NewDirectory(
		"bill", "bill buttlicker", "some-rand-id", filepath.Join(testDir, "bill"),
	)
	tempDrive := svc.NewDrive(
		auth.NewUUID(), "bill", auth.NewUUID(), filepath.Join(testDir, "bill"), tempDir.ID, tempDir,
	)
	tmpUser := auth.NewUser("bill", "bill123", "bill@bill.com", testDir, false)
	return tempDrive, tempDir, tmpUser
}

// creates an empty directory under ../nimbus/pkg/files/testing
func MakeTmpDir(t *testing.T, path string) (*svc.Directory, error) {
	if err := os.Mkdir(path, 0666); err != nil {
		return nil, fmt.Errorf("[ERROR] unable to create temporary directory: %v", err)
	}
	dir := svc.NewDirectory("tmp", "me", "some-rand-id", path)
	return dir, nil
}

// make a bunch of temp .txt files of varying sizes.
// under pkg/files/testing/tmp
func MakeABunchOfTxtFiles(total int) ([]*svc.File, error) {
	tmpDir := filepath.Join(GetTestingDir(), "tmp")

	files := make([]*svc.File, 0)
	for i := 0; i < total; i++ {
		fileName := fmt.Sprintf("tmp-%d.txt", i)
		filePath := filepath.Join(tmpDir, fileName)
		f, err := MakeTmpTxtFile(filePath, RandInt(10000))
		if err != nil {
			return nil, fmt.Errorf("error creating temporary file: %v", err)
		}
		files = append(files, f)
	}
	return files, nil
}

// create a temporary root directory with files and a subdirectory,
// also with files, under testing/tmp
//
// returns complete test root with directory and files
func MakeTmpDirs(t *testing.T) *svc.Directory {
	// make our temporary directory
	d, err := MakeTmpDir(t, filepath.Join(GetTestingDir(), "tmp"))
	if err != nil {
		Fatal(t, err)
	}

	// create some temp files and associated file pointers
	files, err := MakeABunchOfTxtFiles(10)
	if err != nil {
		Fatal(t, err)
	}
	tmpRoot := svc.NewRootDirectory("root", "me", "some-rand-id", filepath.Join(GetTestingDir(), "tmp"))
	tmpRoot.AddFiles(files)

	// add a subdirectory with files so we can test traversal
	sd, err := MakeTmpDir(t, filepath.Join(tmpRoot.Path, "tmpSubDir"))
	if err != nil {
		Fatal(t, err)
	}

	moreFiles := make([]*svc.File, 0)
	for i := 0; i < 10; i++ {
		fname := fmt.Sprintf("tmp-%d.txt", i)
		f, err := MakeTmpTxtFile(filepath.Join(sd.Path, fname), RandInt(1000))
		if err != nil {
			Fatal(t, err)
		}
		moreFiles = append(moreFiles, f)
	}

	sd.AddFiles(moreFiles)
	d.AddSubDir(sd)
	tmpRoot.AddSubDir(d)

	return tmpRoot
}

// clean all contents from the testing directory
func Clean(t *testing.T, dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("[ERROR] could not open test directory: %v", err)
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		t.Errorf("[ERROR] unable to read directory (%s): \n%v\n", dir, err)
	}

	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dir, name)); err != nil {
			t.Errorf("[ERROR] unable to remove file: %v", err)
		}
	}

	return nil
}
