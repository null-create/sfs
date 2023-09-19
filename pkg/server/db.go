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
	c := ServiceConfig()
	dbRoot := filepath.Join(c.S.SvcRoot, "dbs")
	dbPath := filepath.Join(dbRoot, dbName)
	return db.NewQuery(dbPath, false)
}

// get file info from db
func findFile(fileID string, q *db.Query) (*svc.File, error) {
	f, err := q.GetFile(fileID)
	if err != nil {
		return nil, err
	}
	if f == nil { // not found
		return nil, nil
	}
	return f, nil
}

// get user data from db
func findUser(userID string, q *db.Query) (*auth.User, error) {
	u, err := q.GetUser(userID)
	if err != nil {
		return nil, err
	}
	if u == nil { // not found
		return nil, nil
	}
	return u, nil
}

// get directory data from db
func findDir(dirID string, q *db.Query) (*svc.Directory, error) {
	d, err := q.GetDirectory(dirID)
	if err != nil {
		return nil, err
	}
	if d == nil { // not found
		return nil, nil
	}
	return d, nil
}

// get drive data from db
func findDrive(driveID string, q *db.Query) (*svc.Drive, error) {
	d, err := q.GetDrive(driveID)
	if err != nil {
		return nil, err
	}
	if d == nil { // not found
		return nil, nil
	}
	return d, nil
}
