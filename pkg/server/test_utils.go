package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

// at ~1 byte per character, and at 49 characters (inlcuding spaces),
// this string is roughly 49 bytes in size, depending on encoding.
//
// in go's case, this is a utf-8 encoded string, so this is roughly 49 bytes
const txtData string = "all work and no play makes jack a dull boy\n"

// run an individual test as part of a series of larger tests
func RunTestStage(stageName string, test func()) {
	log.Print(fmt.Sprintf("=============== %s Testing Stage ===============", stageName))
	test()
	log.Print("================================================")
}

// build path to test file directory. creates testing directory if it doesn't exist.
func GetTestingDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("[WARNING] unable to find testing directory: %v\ncreating a new one...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0644); err != nil {
			log.Fatalf("unable to create testing directory: %v", err)
		}
	}
	return filepath.Join(curDir, "testing")
}

// clean all contents from the testing directory
func Clean(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dir, name)); err != nil {
			return err
		}
	}
	return nil
}

// like Fatal() but you can specify the directory to clean
func Fail(t *testing.T, dir string, err error) {
	if err2 := Clean(dir); err2 != nil {
		log.Fatal(err2)
	}
	t.Fatalf("[ERROR] %v", err)
}

// handle test failures
//
// calls Clean() followed by t.Fatalf()
// handle test failures
//
// calls Clean() followed by t.Fatalf()
func Fatal(t *testing.T, err error) {
	if err2 := Clean(GetTestingDir()); err2 != nil {
		log.Fatal(err2)
	}
	t.Fatalf("[ERROR] %v", err)
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

	f := svc.NewFile(filepath.Base(filePath), "some-rand-id", filePath)
	for i := 0; i < textReps; i++ {
		_, err := file.WriteString(txtData)
		if err != nil {
			return nil, fmt.Errorf("error writing to test file: %v", err)
		}
	}
	return f, nil
}

// make a bunch of temp .txt files of varying sizes.
// under pkg/files/testing/tmp
func MakeABunchOfTxtFiles(total int, loc string) ([]*svc.File, error) {
	files := make([]*svc.File, 0, total)
	for i := 0; i < total; i++ {
		fileName := fmt.Sprintf("tmp-%d.txt", i)
		filePath := filepath.Join(loc, fileName)

		f, err := MakeTmpTxtFile(filePath, RandInt(1000))
		if err != nil {
			return nil, fmt.Errorf("error creating temporary file: %v", err)
		}

		files = append(files, f)
	}
	return files, nil
}

// ---- tmp dirs

// creates an empty directory under ../nimbus/pkg/files/testing
func MakeTmpDir(t *testing.T, path string) (*svc.Directory, error) {
	if err := os.Mkdir(path, 0666); err != nil {
		return nil, fmt.Errorf("[ERROR] unable to create temporary directory: %v", err)
	}
	dir := svc.NewDirectory("tmp", "me", path)
	return dir, nil
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
	files, err := MakeABunchOfTxtFiles(10, d.Path)
	if err != nil {
		Fatal(t, err)
	}
	tmpRoot := svc.NewRootDirectory("root", "me", filepath.Join(GetTestingDir(), "tmp"))
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

	// build the directories
	sd.AddFiles(moreFiles)
	d.AddSubDir(sd)
	tmpRoot.AddSubDir(d)

	return tmpRoot
}

// make testing drive with test files, directories, and subdirectories.
// all test files will be within the test directory.
func MakeTmpDrive(t *testing.T) *svc.Drive {
	root := MakeTmpDirs(t)
	drive := svc.NewDrive(auth.NewUUID(), "bill buttlicker", root.OwnerID, root.Path, root.ID, root)
	return drive
}

// make a tmp empty drive
func MakeEmptyTmpDrive(t *testing.T) *svc.Drive {
	tmpRoot := svc.NewRootDirectory("tmp", auth.NewUUID(), GetTestingDir())
	testDrv := svc.NewDrive(auth.NewUUID(), "me", tmpRoot.OwnerID, GetTestingDir(), tmpRoot.ID, tmpRoot)
	return testDrv
}
