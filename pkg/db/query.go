package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/sfs/pkg/logger"

	_ "github.com/mattn/go-sqlite3"
)

// Query objects represent connections to to our database.
//
// Setting isSingleton to true will allow Query to automatically
// switch between different databases.
type Query struct {
	DBPath    string         // database directory path
	CurDB     string         // current database we're connecting to
	Debug     bool           // debug flag
	Singleton bool           // flag for whether this is being use as a singleton
	log       *logger.Logger // database logger
	Conn      *sql.DB        // db connection
	Stmt      *sql.Stmt      // SQL statement
	DBs       []string       // list of available databases to both client and server
}

// returns a new query object.
func NewQuery(dbPath string, isSingleton bool) *Query {
	return &Query{
		DBPath:    dbPath,
		CurDB:     "",
		Debug:     false,
		log:       logger.NewLogger("Database", "None"),
		Singleton: isSingleton,
		DBs:       []string{"users", "drives", "directories", "files"},
	}
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
// used only in singleton mode. will be a no-op if/when used
// otherwise.
func (q *Query) WhichDB(dbName string) {
	if q.Singleton && q.DBPath != "" {
		q.CurDB = filepath.Join(q.DBPath, dbName)
	}
}

// connect to a database
//
// must be followed by a defer q.Conn.Close() call when used.
func (q *Query) Connect() error {
	var path string
	if q.Singleton {
		path = q.CurDB
	} else {
		path = q.DBPath
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		q.log.Error(fmt.Sprintf("failed to connect to database: %v", err))
		return fmt.Errorf("failed to open database: %v", err)
	}
	q.Conn = db
	return nil
}

func (q *Query) Close() error {
	if err := q.Conn.Close(); err != nil {
		q.log.Error(fmt.Sprintf("failed to close database connnection: %v", err))
		return fmt.Errorf("unable to close database connection: %v", err)
	}
	return nil
}
