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
const txtData string = "all work and no play makes jack a dull boy"

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
		log.Printf("[WARNING] unable to get testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0644); err != nil {
			log.Fatalf("unable to create test directory: %v", err)
		}
	}
	return filepath.Join(curDir, "testing")
}

func GetStateDir() string {
	curDir, err := os.Getwd()
	if err != nil {
		log.Printf("[WARNING] unable to find state file testing directory: %v\ncreating...", err)
		if err := os.Mkdir(filepath.Join(curDir, "testing"), 0644); err != nil {
			log.Fatalf("unable to create state file testing directory: %v", err)
		}
	}
	return filepath.Join(curDir, "state")
}

// handle test failures
//
// calls Clean() followed by t.Fatalf()
func Fatal(t *testing.T, err error) {
	Clean(GetStateDir())
	Clean(GetTestingDir())
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

	f := svc.NewFile(filepath.Base(filePath), "me", filePath)
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

	files := make([]*svc.File, 0)
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

// make a temporary directory under ../testing/tmp/users (usrsPath)
func MakeDummyUser(usrsPath string, i int) (*auth.User, error) {
	// new user
	svcName := fmt.Sprintf("billbuttlicker-%d", i)
	svcDir := filepath.Join(usrsPath, svcName)
	usrRoot := filepath.Join(svcDir, "root")

	if err := os.Mkdir(svcDir, 0644); err != nil {
		return nil, err
	}

	rt := svc.NewRootDirectory(svcName, "bill buttlicker", usrRoot)
	drv := svc.NewDrive(svc.NewUUID(), svcName, "bill buttlicker", svcDir, rt)

	// gen base files for this user
	GenBaseUserFiles(drv.DriveRoot)

	// create dummy files for this user
	fs, err := MakeABunchOfTxtFiles(RandInt(25), rt.RootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create dummy files: %v", err)
	}
	drv.Root.AddFiles(fs)

	// return user struct
	u := auth.NewUser(
		"bill buttlicker",
		"bill",
		"bill@bill.com",
		drv.ID,
		"",
		false,
	)
	return u, nil
}

func MakeABunchOfUsers(total int, usrsPath string) ([]*auth.User, error) {
	usrs := make([]*auth.User, 0)
	for i := 0; i < total; i++ {
		u, err := MakeDummyUser(usrsPath, i)
		if err != nil {
			return nil, err
		}
		usrs = append(usrs, u)
	}
	return usrs, nil
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
