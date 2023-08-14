package db

// "github.com/mattn/go-sqlite3"
import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// baseline queries
const (
	InsertDataUser  = "INSERT INTO users (message) VALUES (?)"
	InsertDataFile  = "INSERT INTO files (message) VALUES (?)"
	GetUserMesssage = "SELECT message FROM users WHERE id = (?)"
)

type Query struct {
	String  string
	Results *sql.Rows
	Conn    *sql.DB
}

func NewQuery() *Query {
	return &Query{}
}

func (q *Query) ShowQuery() {
	fmt.Printf("Query: %s", q.String)
}

func (q *Query) Connect(dbPath string) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	q.Conn = db
}

// connect to database and send query. Query struct will
// hold the last executed query, assuming it was successfull.
func (q *Query) Ask(query string) {
	rows, err := q.Conn.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	q.String = query
	q.Results = rows
	rows.Close()
}

func (q *Query) Rows() *sql.Rows {
	return q.Results
}

func (q *Query) Close() {
	if err := q.Conn.Close(); err != nil {
		log.Fatalf("[ERROR] unable to close database connection: %v", err)
	}
}
