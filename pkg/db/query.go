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
	CreateUserTable = "CREATE TABLE IF NOT EXISTS users (id INT PRIMARY KEY, message VARCHAR(255))"
	CreateFileTable = "CREATE TABLE IF NOT EXISTS files (id INT PRIMARY KEY, message VARCHAR(255))"
	InsertDataUser  = "INSERT INTO users (message) VALUES (?)"
	InsertDataFile  = "INSERT INTO files (message) VALUES (?)"
	GetUserMesssage = "SELECT message FROM users WHERE id = (?)"
)

type Query struct {
	String  string
	Results *sql.Rows
	Conn    *sql.DB
}

func (q *Query) Connect(dbPath string) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	q.Conn = db
}

func (q *Query) Add(queryStr string) {
	q.String = queryStr
}

func (q *Query) ShowQuery() {
	fmt.Printf("Query: %s", q.String)
}

// connect to database and send query
func (q *Query) Ask() {
	rows, err := q.Conn.Query(q.String)
	if err != nil {
		log.Fatal(err)
	}
	q.Results = rows
	rows.Close()
}

func (q *Query) Close() {
	err := q.Conn.Close()
	if err != nil {
		log.Fatal(err)
	}
}
