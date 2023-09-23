package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// Query objects represent connections to to our database.
//
// if instantiated as a singleton, then the query will prepare
// a map of sql statements that can be used during run time.
type Query struct {
	DBPath string
	Query  string
	Debug  bool

	Conn  *sql.DB              // db connection
	Stmt  *sql.Stmt            // SQL statement (when used as a one time object)
	Stmts map[string]*sql.Stmt // SQL statements (when used as a singleton)
}

// returns a new query struct
func NewQuery(dbPath string, isSingleton bool) *Query {
	q := &Query{
		DBPath: dbPath,
		Query:  "",
		Debug:  false,
	}
	if isSingleton {
		q.Stmts = prepQueries(dbPath)
	}
	return q
}

func prepQueries(dbPath string) map[string]*sql.Stmt {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("unable to open database: %v", err)
	}
	defer conn.Close()

	queries := []string{
		CreateFileTable,
		CreateDirectoryTable,
		CreateDriveTable,
		CreateUserTable,
		DropTableQuery,
		AddFileQuery,
		AddDirQuery,
		AddDriveQuery,
		AddUserQuery,
		FindAllUsersQuery,
		FindAllDrivesQuery,
		FindAllDirsQuery,
		FindAllFilesQuery,
		FindFileQuery,
		FindDirQuery,
		FindDriveQuery,
		FindUserQuery,
	}

	stmts := make(map[string]*sql.Stmt, len(queries))
	for _, query := range queries {
		s, err := conn.Prepare(query)
		if err != nil {
			log.Fatalf("failed to prepare query: %v", err)
		}
		stmts[query] = s
	}
	return stmts
}

// prepare an SQL statement.
func (q *Query) Prepare(query string) error {
	stmt, err := q.Conn.Prepare(query)
	if err != nil {
		return fmt.Errorf("unable to prepare statement: %v", err)
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
