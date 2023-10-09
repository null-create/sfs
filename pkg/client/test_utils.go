package client

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

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
// also with files, under pkg/files/testing/tmp
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

	sd.AddFiles(moreFiles)
	d.AddSubDir(sd)
	tmpRoot.AddSubDir(d)

	return tmpRoot
}

// makes testing/tmp directory objects.
//
// *does not* create actual test directories.
// this is typically done via directory.AddSubDir()
//
// NOTE: none of these test directories have a non-nil parent pointer
func MakeTestDirs(t *testing.T, total int) []*svc.Directory {
	testingDir := GetTestingDir()

	testDirs := make([]*svc.Directory, 0)
	for i := 0; i < total; i++ {
		tdName := fmt.Sprintf("%s%d", TestDirName, i)
		tmpDirPath := filepath.Join(testingDir, tdName)
		testDirs = append(testDirs, svc.NewDirectory(tdName, "me", tmpDirPath))
	}
	return testDirs
}

// ---- files

// "randomly" update some files
// whenever RandInt() returns an even value
func MutateFiles(t *testing.T, files map[string]*svc.File) map[string]*svc.File {
	for _, f := range files {
		if RandInt(100)%2 == 0 {
			var data string
			total := RandInt(5000)
			for i := 0; i < total; i++ {
				data += txtData
			}
			if err := f.Save([]byte(data)); err != nil {
				Fatal(t, err)
			}
		}
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
	f := svc.NewFile(filepath.Base(filePath), "me", filePath)
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

func MakeDummyFiles(t *testing.T, total int) []*svc.File {
	testDir := GetTestingDir()

	// build dummy file objects + test files
	testFiles := make([]*svc.File, 0)
	for i := 0; i < total; i++ {
		tfName := fmt.Sprintf("tmp-%d.txt", i)
		testFiles = append(testFiles, svc.NewFile(tfName, "me", filepath.Join(testDir, tfName)))
	}

	return testFiles
}

// makes temp files and file objects for testing purposes
func MakeTestFiles(t *testing.T, total int) ([]*svc.File, error) {
	testDir := GetTestingDir()

	// build dummy file objects + test files
	testFiles := make([]*svc.File, 0)
	for i := 0; i < total; i++ {
		tfName := fmt.Sprintf("testdoc-%d.txt", i)
		tfPath := filepath.Join(testDir, tfName)

		file, err := os.Create(tfPath)
		if err != nil {
			t.Fatalf("[ERROR] failed to create test file: %v", err)
		}
		file.Write([]byte(txtData))
		file.Close()

		testFiles = append(testFiles, svc.NewFile(tfName, "me", tfPath))
	}
	return testFiles, nil
}

// build large test text files in a specified directory
//
// builds one huge string. really slow.
func MakeLargeTestFiles(total int, dest string) ([]*svc.File, error) {
	testFiles := make([]*svc.File, 0)
	for i := 0; i < total; i++ {
		tfName := fmt.Sprintf("tmpXL-%d.txt", i)
		tfPath := filepath.Join(dest, tfName)

		file, err := os.Create(tfPath)
		if err != nil {
			return nil, err
		}

		var data string
		for i := 0; i < 10000; i++ {
			data += txtData
		}
		if _, err := file.Write([]byte(data)); err != nil {
			return nil, err
		}
		file.Close()

		testFiles = append(testFiles, svc.NewFile(tfName, "me", tfPath))
	}
	return testFiles, nil
}

func AddLargeTestFile() (*svc.File, error) {
	tfDir := filepath.Join(GetWd(), "test_files")
	tf := filepath.Join(tfDir, "me.png")
	dest := filepath.Join(GetTestingDir(), "me.png")
	if err := Copy(tf, dest); err != nil {
		return nil, err
	}
	return svc.NewFile("me.png", "me", tf), nil
}

func RemoveTestFiles(t *testing.T, total int) error {
	testDir := GetTestingDir()

	for i := 0; i < total; i++ {
		tfName := fmt.Sprintf("testdoc-%d.txt", i)
		tfPath := filepath.Join(testDir, tfName)

		if err := os.Remove(tfPath); err != nil {
			return fmt.Errorf("[ERROR] unable to remove test file: %v", err)
		}
	}
	return nil
}

// make test files within a specified directory
func MakeTestDirFiles(t *testing.T, total int, tdPath string) []*svc.File {
	testFiles := make([]*svc.File, 0)

	for i := 0; i < total; i++ {
		name := fmt.Sprintf("test-file-%d.txt", i)
		tfPath := filepath.Join(tdPath, name)

		// Create creates or truncates the named file.
		file, err := os.Create(tfPath)
		if err != nil {
			t.Fatalf("[ERROR] failed to create test file: %v", err)
		}
		file.Write([]byte(txtData))
		file.Close()

		testFiles = append(testFiles, svc.NewFile(name, "me", tfPath))
	}

	return testFiles
}
