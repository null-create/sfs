package client

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sfs/pkg/auth"
	svc "github.com/sfs/pkg/service"
)

//---------- test fixtures & utils --------------------------------

// at ~1 byte per character, and at 49 characters (inlcuding spaces),
// this string is roughly 49 bytes in size, depending on encoding.
//
// in go's case, this is a utf-8 encoded string, so this is roughly 49 bytes
const txtData string = "all work and no play makes jack a dull boy\n"
const TestDirName string = "testDir"

// ---- global

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
		log.Printf("[ERROR] unable to get testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0666); err != nil {
			log.Fatalf("[ERROR] unable to create test directory: %v", err)
		}
	}
	return filepath.Join(curDir, "testing")
}

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

// handle test failures
//
// calls Clean() followed by t.Fatalf()
func Fatal(t *testing.T, err error) {
	if err := Clean(t, GetTestingDir()); err != nil {
		log.Fatal(err)
	}
	t.Fatalf("[ERROR] %v", err)
}

// clean all contents from the testing directory
func Clean(t *testing.T, dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		t.Errorf("[ERROR] unable to read directory: %v", err)
	}

	for _, name := range names {
		if err = os.RemoveAll(filepath.Join(dir, name)); err != nil {
			t.Errorf("[ERROR] unable to remove file: %v", err)
		}
	}

	return nil
}

// ----- misc.

// for random file names
func randString(length int) string {
	rand.Seed(time.Now().UnixNano())
	charSet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	randbytes := make([]byte, length)
	for i := 0; i < length; i++ {
		randbytes[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(randbytes)
}

// ---- tmp dirs

// creates an empty directory under ../nimbus/pkg/files/testing
func MakeTmpDir(t *testing.T, path string) (*svc.Directory, error) {
	if err := os.Mkdir(path, 0666); err != nil {
		return nil, fmt.Errorf("[ERROR] unable to create temporary directory: %v", err)
	}
	dir := svc.NewDirectory("tmp", "me", "some-rand-id", path)
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

// create a temporary root directory with files and a subdirectory,
// also with files, under a specified path.
//
// returns complete test root with directory and files
func MakeTmpDirsWithPath(t *testing.T, path string, driveID string) *svc.Directory {
	// make our temporary directory
	d, err := MakeTmpDir(t, filepath.Join(path, "tmp"))
	if err != nil {
		Fatal(t, err)
	}

	// create tmp service root directory with some files and
	// make sure we properly assign these files and directories to the temp drivea
	tmpRoot := svc.NewRootDirectory("root", "some-rand-id", driveID, filepath.Join(path, "tmp"))
	files, err := MakeABunchOfTxtFilesWithLoc(10, d.Path)
	if err != nil {
		Fatal(t, err)
	}
	for _, f := range files {
		f.DriveID = driveID
	}
	tmpRoot.AddFiles(files)

	// add a subdirectory also with files
	sd, err := MakeTmpDir(t, filepath.Join(tmpRoot.Path, "tmpSubDir"))
	if err != nil {
		Fatal(t, err)
	}
	sd.DriveID = driveID
	moreFiles, err := MakeABunchOfTxtFilesWithLoc(10, sd.Path)
	if err != nil {
		Fatal(t, err)
	}
	for _, f := range moreFiles {
		f.DriveID = driveID
	}

	// build the directory structure
	sd.AddFiles(moreFiles)
	d.AddSubDir(sd)
	tmpRoot.AddSubDir(d)

	return tmpRoot
}

// ---- files

// alter a test file with additional data
func MutateFile(t *testing.T, f *svc.File) {
	var data string
	total := RandInt(5000)
	for i := 0; i < total; i++ {
		data += txtData
	}
	if err := f.Save([]byte(data)); err != nil {
		Fail(t, GetTestingDir(), err)
	}
}

// "randomly" update some files
// whenever RandInt() returns an even value
//
// NOTE: could potentially cause a test to be flakey!
// this is because sometimes randint will pick 1 and if that's the case
// then nothing will be modified, and no potential updates could be
// caught by dir.WalkU().
func MutateFiles(t *testing.T, files map[string]*svc.File) map[string]*svc.File {
	var mutated int
	for _, f := range files {
		if RandInt(1000)%2 == 0 {
			var data string
			total := RandInt(5000)
			for i := 0; i < total; i++ {
				data += txtData
			}
			if err := f.Save([]byte(data)); err != nil {
				Fatal(t, err)
			}
			mutated += 1
		}
	}
	if mutated == 0 {
		log.Print("[WARNING] no test files altered!")
	} else {
		log.Printf("[INFO] altered %d files", mutated)
	}
	return files
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

// make a bunch of temp .txt files of varying sizes.
// under pkg/files/testing/tmp
func MakeABunchOfTxtFilesWithLoc(total int, loc string) ([]*svc.File, error) {
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

// ---- test drives and stuff

// make testing drive with test files, directories, and subdirectories.
// all test files will be within the test directory.
func MakeTmpDrive(t *testing.T) *svc.Drive {
	root := MakeTmpDirs(t)
	drive := svc.NewDrive(auth.NewUUID(), "bill buttlicker", root.OwnerID, root.Path, root.ID, root)
	return drive
}

// make testing drive with test files, directories, and subdirectories
// under a specified path.
func MakeTmpDriveWithPath(t *testing.T, path string) *svc.Drive {
	tmpDriveID := auth.NewUUID()
	root := MakeTmpDirsWithPath(t, path, tmpDriveID)
	drive := svc.NewDrive(tmpDriveID, "bill buttlicker", root.OwnerID, root.Path, root.ID, root)
	return drive
}
