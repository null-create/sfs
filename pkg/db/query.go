package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Query struct {
	DBPath string
	Query  string
	Debug  bool

	Conn *sql.DB   // db connection
	Stmt *sql.Stmt // SQL statement
}

// returns a new query struct
func NewQuery(dbPath string) *Query {
	return &Query{
		DBPath: dbPath,
	}
}

// prepare a statement. must be followed by q.Stmt.Close() when called!
func (q *Query) Prepare(query string) error {
	stmt, err := q.Conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("[ERROR] unable to prepare statement: %v", err)
	}
	q.Query = query
	q.Stmt = stmt
	return nil
}

// connect to a database given the assigned dbPath when query was initialized
//
// must be followed by a defer q.Conn.Close() statement when called!
func (q *Query) Connect() error {
	if q.DBPath == "" {
		return fmt.Errorf("no DB path specified")
	}
	db, err := sql.Open("sqlite3", q.DBPath)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
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
