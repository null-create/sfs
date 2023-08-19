package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// all database files must end with .db!
func NewDB(dbName string, pathToNewDB string) {
	if dbName == "users" {
		NewTable(pathToNewDB, CreateUserTable)
	} else if dbName == "drives" {
		NewTable(pathToNewDB, CreateDriveTable)
	} else if dbName == "directories" {
		NewTable(pathToNewDB, CreateDirectoryTable)
	} else if dbName == "files" {
		NewTable(pathToNewDB, CreateFileTable)
	} else {
		log.Printf("[DEBUG] unknown database category: %s", dbName)
	}
}

// create a new table with the given query
func NewTable(path string, query string) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Fatalf("[ERROR] unable to open database: \n%v\n", err)
	}
	defer db.Close()

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("[ERROR] failed to create table: \n%v\n", err)
	}
}
