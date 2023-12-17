package db

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Query objects represent connections to to our database.
//
// If instantiated as a singleton, then the query will prepare
// a map of sql statements that can be used during run time.
type Query struct {
	DBPath string // database directory path
	CurDB  string // current database we're connecting to
	Debug  bool   // debug flag

	Singleton bool // flag for whether this is being use as a singleton

	Conn *sql.DB   // db connection
	Stmt *sql.Stmt // SQL statement
	DBs  []string  // list of available databases to both client and server
}

// returns a new query object.
func NewQuery(dbPath string, isSingleton bool) *Query {
	dbs := []string{"users", "drives", "directories", "files"}
	return &Query{
		DBPath:    dbPath,
		CurDB:     "",
		Debug:     false,
		Singleton: isSingleton,
		DBs:       dbs,
	}
}

// prepare an SQL statement.
func (q *Query) Prepare(query string) error {
	stmt, err := q.Conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("unable to prepare statement: %v", err)
	}
	q.Stmt = stmt
	return nil
}

// make sure this is a valid database
func (q *Query) ValidDb(dbName string) bool {
	for _, db := range q.DBs {
		if dbName == db {
			return true
		}
	}
	return false
}

// sets the file path to the db we want to connect to.
//
// used only in singleton mode. will be a no-op when used
// otherwise
func (q *Query) WhichDB(dbName string) {
	if q.Singleton && q.DBPath != "" {
		q.CurDB = filepath.Join(q.DBPath, dbName)
	} else {
		log.Print("[WARNING] q.DBpath was not set")
	}
}

// connect to a database
//
// must be followed by a defer q.Conn.Close() statement when called!
func (q *Query) Connect() error {
	var path string
	if q.Singleton {
		path = q.CurDB
	} else {
		path = q.DBPath
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	q.Conn = db
	return nil
}

func (q *Query) Close() error {
	if err := q.Conn.Close(); err != nil {
		return fmt.Errorf("unable to close database connection: %v", err)
	}
	return nil
}
