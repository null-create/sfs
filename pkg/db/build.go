package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

// all database files must end with .db!
func NewDB(dbName string, pathToNewDB string) {
	if dbName == "users" {
		new(pathToNewDB, CreateUserTable)
	} else if dbName == "drives" {
		new(pathToNewDB, CreateDriveTable)
	} else if dbName == "directories" {
		new(pathToNewDB, CreateDirectoryTable)
	} else if dbName == "files" {
		new(pathToNewDB, CreateFileTable)
	} else {
		log.Printf("[DEBUG] unknown database category: %s", dbName)
	}
}

// build a temp testing directory using a specified SQL query
func NewTestDB(dbName string, pathToNewDB string, query string) {
	db, err := sql.Open("sqlite3", pathToNewDB)
	if err != nil {
		log.Fatalf("[ERROR] unable to open database: \n%v\n", err)
	}
	defer db.Close()

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("[ERROR] failed to create table: \n%v\n", err)
	}
}

func new(path string, query string) {
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
