package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

const (
	CreateUserTable = `
	IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'Users')
	BEGIN
			CREATE TABLE UserTable (
					id VARCHAR(50) PRIMARY KEY,
					name VARCHAR(255),
					username VARCHAR(50),
					email VARCHAR(255),
					password VARCHAR(100),
					last_login_date DATETIME,
					is_admin BIT,
					total_files INT,
					total_directories INT
			);
	END`

	CreateFileTable = `
	IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'Files')
	BEGIN
			CREATE TABLE Files (
					id VARCHAR(50) PRIMARY KEY,
					name VARCHAR(255),
					owner VARCHAR(50),
					protected BIT,
					key VARCHAR(100),
					path VARCHAR(255),
					server_path VARCHAR(255),
					client_path VARCHAR(255),
					checksum VARCHAR(255),
					algorithm VARCHAR(50)
			);
	END
	`
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
