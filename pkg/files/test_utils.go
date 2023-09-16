package files

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
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
	Clean(t, GetTestingDir())
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
func MakeTmpDir(t *testing.T, path string) (*Directory, error) {
	if err := os.Mkdir(path, 0666); err != nil {
		return nil, fmt.Errorf("[ERROR] unable to create temporary directory: %v", err)
	}
	dir := NewDirectory("tmp", "me", path)
	return dir, nil
}

// create a temporary root directory with files and a subdirectory,
// also with files, under pkg/files/testing/tmp
func MakeTmpDirs(t *testing.T) *Directory {
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
	tmpRoot := NewRootDirectory("root", "me", filepath.Join(GetTestingDir(), "tmp"))
	tmpRoot.AddFiles(files)

	// add a subdirectory with files so we can test traversal
	sd, err := MakeTmpDir(t, filepath.Join(tmpRoot.Path, "tmpSubDir"))
	if err != nil {
		Fatal(t, err)
	}

	moreFiles := make([]*File, 0)
	for i := 0; i < 10; i++ {
		fname := fmt.Sprintf("tmp-%d.txt", i)
		f, err := MakeTmpTxtFile(filepath.Join(sd.Path, fname), RandInt(1000))
		if err != nil {
			Fatal(t, err)
		}
		moreFiles = append(moreFiles, f)
	}

	sd.AddFiles(moreFiles)
	d.addSubDir(sd)
	tmpRoot.addSubDir(d)

	return tmpRoot
}

// ---- files

// randomly append test data to give each
// file a roughly unique size
func MutateFiles(t *testing.T, files map[string]*File) map[string]*File {
	for _, f := range files {
		if RandInt(2) == 1 {
			var data string
			total := RandInt(1000)
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

func MakeDummyFiles(t *testing.T, total int) []*File {
	testDir := GetTestingDir()

	// build dummy file objects + test files
	testFiles := make([]*File, 0)
	for i := 0; i < total; i++ {
		tfName := fmt.Sprintf("tmp-%d.txt", i)
		testFiles = append(testFiles, NewFile(tfName, "me", filepath.Join(testDir, tfName)))
	}

	return testFiles
}

// makes temp files and file objects for testing purposes
func MakeTestFiles(t *testing.T, total int) ([]*File, error) {
	testDir := GetTestingDir()

	// build dummy file objects + test files
	testFiles := make([]*File, 0)
	for i := 0; i < total; i++ {
		tfName := fmt.Sprintf("testdoc-%d.txt", i)
		tfPath := filepath.Join(testDir, tfName)

		file, err := os.Create(tfPath)
		if err != nil {
			t.Fatalf("[ERROR] failed to create test file: %v", err)
		}
		file.Write([]byte(txtData))
		file.Close()

		testFiles = append(testFiles, NewFile(tfName, "me", tfPath))
	}
	return testFiles, nil
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
func MakeTestDirFiles(t *testing.T, total int, tdPath string) []*File {
	testFiles := make([]*File, 0)

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

		testFiles = append(testFiles, NewFile(name, "me", tfPath))
	}

	return testFiles
}

// makes testing/tmp directory objects.
//
// *does not* create actual test directories.
// this is typically done via directory.AddSubDir()
func MakeTestDirs(t *testing.T, total int) []*Directory {
	testingDir := GetTestingDir()

	testDirs := make([]*Directory, 0)
	for i := 0; i < total; i++ {
		tdName := fmt.Sprintf("%s%d", TestDirName, i)
		tmpDirPath := filepath.Join(testingDir, tdName)

		// NOTE: none of these test directories have a non-nil parent pointer
		testDirs = append(testDirs, NewDirectory(tdName, "me", tmpDirPath))
	}
	return testDirs
}

// make a temp .txt file of n size (in bytes).
//
// n is determined by textReps since that will be how
// many times testData is written to the text file
//
// returns a file pointer to the new temp file
func MakeTmpTxtFile(filePath string, textReps int) (*File, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	f := NewFile(filepath.Base(filePath), "me", filePath)
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
func MakeABunchOfTxtFiles(total int) ([]*File, error) {
	tmpDir := filepath.Join(GetTestingDir(), "tmp")

	files := make([]*File, 0)
	for i := 0; i < total; i++ {
		fileName := fmt.Sprintf("tmp-%d.txt", i)
		filePath := filepath.Join(tmpDir, fileName)
		f, err := MakeTmpTxtFile(filePath, RandInt(1000))
		if err != nil {
			return nil, fmt.Errorf("error creating temporary file: %v", err)
		}

		files = append(files, f)
	}
	return files, nil
}
