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
	Debug  bool

	Conn  *sql.DB     // db connection
	Stmt  *sql.Stmt   // SQL statement
	Stmts []*sql.Stmt // SQL statements (when used as a singleton)
}

// returns a new query struct
func NewQuery(dbPath string, isSingleton bool) *Query {
	if isSingleton {
		stmts := prepQueries(dbPath)
		return &Query{
			DBPath: dbPath,
			Debug:  false,
			Query:  "",
			Stmts:  stmts,
		}
	}
	return &Query{
		DBPath: dbPath,
		Query:  "",
		Debug:  false,
	}
}

func prepQueries(dbPath string) []*sql.Stmt {
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("[ERROR] unable to open database: %v", err)
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

	stmts := make([]*sql.Stmt, 0)
	for _, query := range queries {
		s, err := conn.Prepare(query)
		if err != nil {
			log.Fatalf("failed to prepare query: %v", err)
		}
		stmts = append(stmts, s)
	}

	return stmts
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
