package server

import (
	"path/filepath"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
	svc "github.com/sfs/pkg/service"
)

// ----- db utils --------------------------------

// get a one-off db connection to a given db
func getDBConn(dbName string) *db.Query {
	cfg := ServiceConfig()
	return db.NewQuery(filepath.Join(cfg.SvcRoot, "dbs", dbName), false)
}

// get file info from db. file will be nil if not found.
func findFile(fileID string, q *db.Query) (*svc.File, error) {
	f, err := q.GetFile(fileID)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// get a slice of all files for this user. returns an empty slice if none are found.
func getAllFiles(userID string, q *db.Query) ([]*svc.File, error) {
	files, err := q.GetUsersFiles(userID)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// get user data from db. user will be nil if not found
func findUser(userID string, q *db.Query) (*auth.User, error) {
	u, err := q.GetUser(userID)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// get directory data from db. dir will be nil if not found.
func findDir(dirID string, q *db.Query) (*svc.Directory, error) {
	d, err := q.GetDirectory(dirID)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// get drive data from db. drive will be nil if not found.
func findDrive(driveID string, q *db.Query) (*svc.Drive, error) {
	d, err := q.GetDrive(driveID)
	if err != nil {
		return nil, err
	}
	return d, nil
}
