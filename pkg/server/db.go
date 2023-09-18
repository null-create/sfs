package server

import (
	"fmt"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	svc "github.com/sfs/pkg/service"
)

// ----- db utils --------------------------------

// get file info from db
func findFile(fileID string, q *db.Query) (*svc.File, error) {
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
func findUser(userID string, q *db.Query) (*auth.User, error) {
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
func findDir(dirID string, q *db.Query) (*svc.Directory, error) {
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
func findDrive(driveID string, q *db.Query) (*svc.Drive, error) {
	d, err := q.GetDrive(driveID)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, fmt.Errorf("no drive found with ID %s", driveID)
	}
	return d, nil
}
