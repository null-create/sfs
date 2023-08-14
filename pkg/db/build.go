package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	CreateUserTable = "CREATE TABLE IF NOT EXISTS users (id VARCHAR(255) PRIMARY KEY, name VARCHAR(255), user_name VARCHAR(255), email VARCHAR(255), password VARCHAR(255))"
	CreateFileTable = "CREATE TABLE IF NOT EXISTS files (id VARCHAR(255) PRIMARY KEY, file_name VARCHAR(255), file_path VARCHAR(255), directory VARCHAR(255), last_sync VARCHAR(255))"
)

// MakeNewDatabase creates a new database during initial set up of
// the file sync service.
func NewUserDB(pathToNewDb string) {
	db, err := sql.Open("sqlite3", pathToNewDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// TODO: standardize db creation query parameters/pattern
	// Create an initial table in the database
	result, err := db.Exec(CreateUserTable)
	if err != nil {
		log.Fatalf("[ERROR] failed to create user table: \n%v\n", err)
	}

	fmt.Printf("[DEBUG] user database created successfully! \n%v\n", result)
}

// MakeNewDatabase creates a new database during initial set up of
// the file sync service.
func NewFileDB(pathToNewDb string) {
	db, err := sql.Open("sqlite3", pathToNewDb)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// TODO: add creation date and way to store db names persistently

	// TODO: standardize db creation query parameters/pattern
	// Create an initial table in the database
	result, err := db.Exec(CreateFileTable)
	if err != nil {
		log.Fatalf("[ERROR] failed to create file table: \n%v\n", err)
	}

	fmt.Printf("[DEBUG] file database created successfully! \n%v\n", result)
}
