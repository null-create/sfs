package server

import (
	"fmt"
	"path/filepath"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	svc "github.com/sfs/pkg/service"
)

// ----- db utils --------------------------------

// get file info from db
func findFile(fileID string, dbDir string) (*svc.File, error) {
	q := db.NewQuery(filepath.Join(dbDir, "files"), false)
	f, err := q.GetFile(fileID)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, fmt.Errorf("no file found with ID %s", fileID)
	}
	return f, nil
}

// get user data from db
func findUser(userID string, dbDir string) (*auth.User, error) {
	q := db.NewQuery(filepath.Join(dbDir, "users"), false)
	u, err := q.GetUser(userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, fmt.Errorf("no user found with ID %s", userID)
	}
	return u, nil
}

// get directory data from db
func findDir(dirID string, dbDir string) (*svc.Directory, error) {
	q := db.NewQuery(filepath.Join(dbDir, "directories"), false)
	d, err := q.GetDirectory(dirID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, fmt.Errorf("no directory found with ID %s", dirID)
	}
	return d, nil
}

// get drive data from db
func findDrive(driveID string, dbDir string) (*svc.Drive, error) {
	q := db.NewQuery(filepath.Join(dbDir, "drives"), false)
	d, err := q.GetDrive(driveID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, fmt.Errorf("no drive found with ID %s", driveID)
	}
	return d, nil
}
