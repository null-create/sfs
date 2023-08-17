package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type Query struct {
	DBPath string
	Query  string

	Conn *sql.DB   // db connection
	Stmt *sql.Stmt // SQL statement
}

// returns a new query struct
// TODO: initialize prepared SQL statements?
// see: https://go.dev/doc/database/prepared-statements
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
	q.Stmt = stmt
	return nil
}

func (q *Query) Connect() {
	db, err := sql.Open("sqlite3", q.DBPath)
	if err != nil {
		log.Fatalf("[ERROR] failed to connect to database: %v", err)
	}
	q.Conn = db
}

func (q *Query) Close() {
	if err := q.Conn.Close(); err != nil {
		log.Fatalf("[ERROR] unable to close database connection: %v", err)
	}
}
