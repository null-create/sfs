package server

import (
	"path/filepath"

	"github.com/sfs/pkg/auth"
	"github.com/sfs/pkg/db"
)

// ----- db utils --------------------------------

// get a one-off db connection to a given db
func getDBConn(dbName string) *db.Query {
	return db.NewQuery(filepath.Join(svcCfg.SvcRoot, "dbs", dbName), false)
}

// get user data from db. user will be nil if not found
func findUser(userID string, q *db.Query) (*auth.User, error) {
	u, err := q.GetUser(userID)
	if err != nil {
		return nil, err
	}
	return u, nil
}
